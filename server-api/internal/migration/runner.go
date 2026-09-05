package migration

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"

	driverMySQL "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var migrationFileName = regexp.MustCompile(`^\d{8}_\d{4}_[a-z0-9][a-z0-9_-]*\.sql$`)

type appliedMigration struct {
	Filename string `gorm:"column:filename"`
	Checksum string `gorm:"column:checksum"`
}

// RunPending 按文件名顺序执行未记录的日期化 SQL，并以文件名和校验和防止重复或篡改执行。
func RunPending(db *gorm.DB, dsn, directory, initFile string) (int, error) {
	if db == nil {
		return 0, errors.New("mysql database is nil")
	}
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename VARCHAR(255) NOT NULL PRIMARY KEY COMMENT '迁移文件名',
			checksum CHAR(64) NOT NULL COMMENT '文件SHA-256校验和',
			applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '执行时间'
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SQL迁移执行记录'`).Error; err != nil {
		return 0, fmt.Errorf("创建迁移记录表失败: %w", err)
	}

	var applied []appliedMigration
	if err := db.Table("schema_migrations").Find(&applied).Error; err != nil {
		return 0, fmt.Errorf("读取迁移记录失败: %w", err)
	}
	appliedByName := make(map[string]string, len(applied))
	for _, item := range applied {
		appliedByName[item.Filename] = item.Checksum
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return 0, fmt.Errorf("读取迁移目录失败: %w", err)
	}
	initBaseName := filepath.Base(initFile)
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		if entry.Name() == initBaseName {
			continue
		}
		if !migrationFileName.MatchString(entry.Name()) {
			return 0, fmt.Errorf("迁移 SQL 文件名不符合 YYYYMMDD_序号_说明.sql 规则: %s", entry.Name())
		}
		files = append(files, entry.Name())
	}
	sort.Strings(files)

	executed := 0
	for _, name := range files {
		path := filepath.Join(directory, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return executed, fmt.Errorf("读取迁移 SQL 失败 %s: %w", name, err)
		}
		checksum := fmt.Sprintf("%x", sha256.Sum256(content))
		if oldChecksum, ok := appliedByName[name]; ok {
			if oldChecksum != checksum {
				return executed, fmt.Errorf("已执行的迁移 SQL 不可修改: %s", name)
			}
			continue
		}
		if err := executeSQLFile(dsn, path); err != nil {
			return executed, fmt.Errorf("执行迁移 SQL 失败 %s: %w", name, err)
		}
		if err := db.Exec("INSERT INTO schema_migrations (filename, checksum) VALUES (?, ?)", name, checksum).Error; err != nil {
			return executed, fmt.Errorf("记录迁移结果失败 %s: %w", name, err)
		}
		executed++
	}
	return executed, nil
}

func executeSQLFile(dsn, path string) error {
	conf, err := driverMySQL.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("解析 MySQL DSN 失败: %w", err)
	}
	if conf.DBName == "" || conf.Net != "tcp" {
		return errors.New("迁移 SQL 仅支持指定数据库名的 tcp MySQL DSN")
	}
	host, port, err := net.SplitHostPort(conf.Addr)
	if err != nil {
		return fmt.Errorf("解析 MySQL 地址失败: %w", err)
	}
	if host == "" || port == "" {
		return errors.New("解析 MySQL 地址失败: 地址不能为空")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	cmd := exec.Command("mysql", "--host="+host, "--port="+port, "--user="+conf.User, "--database="+conf.DBName, "--default-character-set=utf8mb4")
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conf.Passwd)
	cmd.Stdin = file
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}
	return nil
}
