// migrate 命令将当前数据库导出为初始化 SQL。
package main

import (
	"flag"
	"fmt"
	"megin/internal/config"
	"megin/internal/migration"
	"os"
)

var (
	env    = flag.String("env", "dev", "dev/test/prod")
	output = flag.String("output", "sql/20260904_0001_init.sql", "导出的 SQL 文件路径")
)

func main() {
	flag.Parse()

	conf := config.InitConfig(config.GetConfigPath(*env), config.RunModeMixed)
	if conf.Database.Driver != "mysql" {
		fail(fmt.Errorf("migrate 仅支持 mysql，当前驱动为 %q", conf.Database.Driver))
	}
	if err := migration.Dump(conf.Database.Dsn, *output); err != nil {
		fail(err)
	}
	fmt.Printf("迁移完成，初始化 SQL 已更新：%s\n", *output)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "migrate failed:", err)
	os.Exit(1)
}
