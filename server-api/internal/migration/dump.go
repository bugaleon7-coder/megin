// Package migration 提供数据库初始化 SQL 的导出能力。
package migration

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	driverMySQL "github.com/go-sql-driver/mysql"
)

// Dump 将 MySQL 数据库完整导出到 outputPath。只有导出成功后才替换目标文件。
func Dump(dsn, outputPath string) error {
	conf, err := driverMySQL.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("解析 MySQL DSN 失败: %w", err)
	}
	if conf.DBName == "" {
		return errors.New("MySQL DSN 未指定数据库名")
	}
	if conf.Net != "tcp" {
		return fmt.Errorf("migrate 仅支持 tcp MySQL 连接，当前网络类型为 %q", conf.Net)
	}

	host, port, err := net.SplitHostPort(conf.Addr)
	if err != nil {
		return fmt.Errorf("解析 MySQL 地址失败: %w", err)
	}
	if host == "" || port == "" {
		return errors.New("MySQL 地址必须包含主机和端口")
	}

	absOutput, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("解析 SQL 输出路径失败: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(absOutput), 0o755); err != nil {
		return fmt.Errorf("创建 SQL 输出目录失败: %w", err)
	}
	tempFile, err := os.CreateTemp(filepath.Dir(absOutput), "."+filepath.Base(absOutput)+".*.tmp")
	if err != nil {
		return fmt.Errorf("创建临时 SQL 文件失败: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	args := []string{
		"--host=" + host,
		"--port=" + port,
		"--user=" + conf.User,
		"--databases",
		"--skip-add-drop-table",
		"--single-transaction",
		"--routines",
		"--triggers",
		"--events",
		"--set-gtid-purged=OFF",
		"--no-tablespaces",
		"--default-character-set=utf8mb4",
		conf.DBName,
	}
	cmd := exec.Command("mysqldump", args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conf.Passwd)
	cmd.Stdout = tempFile
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("执行 mysqldump 失败: %w: %s", err, stderr.String())
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("关闭临时 SQL 文件失败: %w", err)
	}
	if err := os.Chmod(tempPath, 0o644); err != nil {
		return fmt.Errorf("设置 SQL 文件权限失败: %w", err)
	}
	if err := os.Rename(tempPath, absOutput); err != nil {
		return fmt.Errorf("更新 SQL 文件失败: %w", err)
	}
	return nil
}

// DatabaseExists 检查 DSN 指定的数据库是否已创建。
func DatabaseExists(dsn string) (bool, error) {
	conf, err := driverMySQL.ParseDSN(dsn)
	if err != nil {
		return false, fmt.Errorf("解析 MySQL DSN 失败: %w", err)
	}
	if conf.DBName == "" {
		return false, errors.New("MySQL DSN 未指定数据库名")
	}
	if conf.Net != "tcp" {
		return false, fmt.Errorf("migrate 仅支持 tcp MySQL 连接，当前网络类型为 %q", conf.Net)
	}
	host, port, err := net.SplitHostPort(conf.Addr)
	if err != nil {
		return false, fmt.Errorf("解析 MySQL 地址失败: %w", err)
	}
	if host == "" || port == "" {
		return false, errors.New("MySQL 地址必须包含主机和端口")
	}

	query := "SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = '" + strings.ReplaceAll(conf.DBName, "'", "''") + "'"
	cmd := exec.Command("mysql", "--host="+host, "--port="+port, "--user="+conf.User, "--batch", "--skip-column-names", "--execute="+query)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conf.Passwd)
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("检查数据库是否存在失败: %w", err)
	}
	return strings.TrimSpace(string(output)) == "1", nil
}

// ApplyInit 执行首次安装的初始化 SQL。该文件只建库、建表和导入数据，不删除已有数据库或表。
func ApplyInit(dsn, inputPath string) error {
	conf, err := driverMySQL.ParseDSN(dsn)
	if err != nil {
		return fmt.Errorf("解析 MySQL DSN 失败: %w", err)
	}
	if conf.Net != "tcp" {
		return fmt.Errorf("migrate 仅支持 tcp MySQL 连接，当前网络类型为 %q", conf.Net)
	}
	host, port, err := net.SplitHostPort(conf.Addr)
	if err != nil {
		return fmt.Errorf("解析 MySQL 地址失败: %w", err)
	}
	if host == "" || port == "" {
		return errors.New("MySQL 地址必须包含主机和端口")
	}

	initSQL, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("打开初始化 SQL 文件失败: %w", err)
	}

	cmd := exec.Command("mysql", "--host="+host, "--port="+port, "--user="+conf.User, "--default-character-set=utf8mb4")
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+conf.Passwd)
	// 初始化快照来源于开发库。执行前仅替换建库和选库语句中的库名，
	// 使同一份初始化数据可用于 test、prod 等不同配置库。
	cmd.Stdin = strings.NewReader(rewriteInitDatabase(string(initSQL), conf.DBName))
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行初始化 SQL 失败: %w: %s", err, stderr.String())
	}
	return nil
}

func rewriteInitDatabase(sql, databaseName string) string {
	escapedName := strings.ReplaceAll(databaseName, "`", "``")
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "CREATE DATABASE "),
			strings.HasPrefix(trimmed, "USE `"):
			lines[i] = replaceQuotedDatabaseName(line, escapedName)
		}
	}
	return strings.Join(lines, "\n")
}

func replaceQuotedDatabaseName(line, databaseName string) string {
	prefix, remaining, found := strings.Cut(line, "`")
	if !found {
		return line
	}
	_, suffix, found := strings.Cut(remaining, "`")
	if !found {
		return line
	}
	return prefix + "`" + databaseName + "`" + suffix
}
