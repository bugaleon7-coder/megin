// migrate 命令执行未执行的增量 SQL；--snapshot 时导出初始化 SQL 快照。
package main

import (
	"flag"
	"fmt"
	"megin/internal/config"
	"megin/internal/migration"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	env      = flag.String("env", "dev", "dev/test/prod")
	snapshot = flag.Bool("snapshot", false, "导出当前数据库为初始化 SQL 快照")
)

func main() {
	flag.Parse()

	conf := config.InitConfig(config.GetConfigPath(*env), config.RunModeMixed)
	if conf.Database.Driver != "mysql" {
		fail(fmt.Errorf("migrate 仅支持 mysql，当前驱动为 %q", conf.Database.Driver))
	}
	exists, err := migration.DatabaseExists(conf.Database.Dsn)
	if err != nil {
		fail(err)
	}
	if !exists {
		if err := migration.ApplyInit(conf.Database.Dsn, conf.Migrate.OutputFile()); err != nil {
			fail(err)
		}
	}
	if *snapshot {
		if err := migration.Dump(conf.Database.Dsn, conf.Migrate.OutputFile()); err != nil {
			fail(err)
		}
		fmt.Printf("初始化 SQL 快照已更新：%s\n", conf.Migrate.OutputFile())
		return
	}
	db, err := gorm.Open(mysql.Open(conf.Database.Dsn), &gorm.Config{SkipDefaultTransaction: conf.Database.SkipDefaultTransaction})
	if err != nil {
		fail(fmt.Errorf("连接 MySQL 失败: %w", err))
	}
	count, err := migration.RunPending(db, conf.Database.Dsn, conf.Migrate.Directory(), conf.Migrate.OutputFile())
	if err != nil {
		fail(err)
	}
	fmt.Printf("迁移完成，执行 %d 个增量 SQL 文件\n", count)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "migrate failed:", err)
	os.Exit(1)
}
