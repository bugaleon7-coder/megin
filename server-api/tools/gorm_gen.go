package main

import (
	"megin/pkg/gen"
	"megin/tools/tpl"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type RepoTplData struct {
	ModelName string
}

// https://gorm.io/gen/database_to_structs.html

func main() {
	//要生成的表名
	tableNames := []string{
		"command_execution_records",
	}

	g := gen.NewGenerator(gen.Config{
		//路径
		BasePkgPath: "./internal/api",
		ModelPath:   "./internal/api/model",
		RepoPath:    "./internal/api/repository",
		ServicePath: "./internal/api/service",
		HandlerPath: "./internal/api/handler",

		//是否覆盖已存在
		EnableModelOverride:   true,
		EnableRepoOverride:    false,
		EnableServiceOverride: false,
		EnableHandleOverride:  false,

		//是否启用
		EnableModel:   true,
		EnableRepo:    true,
		EnableService: true,
		EnableHandler: true,

		Mode: gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		//FieldNullable:  true,
	})

	g.WithFileNameStrategy(func(tableName string) (fileName string) {
		return tableName
	})

	gormdb, _ := gorm.Open(mysql.Open("root:111111@tcp(127.0.0.1:3306)/xauto?charset=utf8mb4&parseTime=True&loc=Local"))
	g.UseDB(gormdb) // reuse your gorm db

	modelOpts := []gen.ModelOpt{
		//gen.FieldIgnore("address"),
		gen.WithMethod(tpl.CommonMethod{}),
	}

	// 生成所有表的代码
	for _, tableName := range tableNames {
		g.GenerateModel(tableName, modelOpts...)
	}
	g.Execute()
}
