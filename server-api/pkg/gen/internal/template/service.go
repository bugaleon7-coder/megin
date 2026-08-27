package template

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Service = NotEditMark + `
package service

import (
	"megin/internal/repository"
	"megin/pkg/context/api"
)

// {{.StructComment}}
type {{.ModelStructName}} struct {
	Service
	Repo *repo.{{.ModelStructName}}
}

func New{{.ModelStructName}}(ctx *api.Context) *{{.ModelStructName}} {
	service := {{.ModelStructName}}{}
	service.initialize(ctx)
	service.Repo = repo.New{{.ModelStructName}}(ctx)
	return &service
}

`
