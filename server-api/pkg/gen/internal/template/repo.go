package template

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Repo = NotEditMark + `
package repo

import (
	"megin/internal/model"
	"megin/pkg/context/api"
)

// {{.StructComment}}
type {{.ModelStructName}} struct {
	Repository[model.{{.ModelStructName}}]
}

func New{{.ModelStructName}}(ctx *api.Context) *{{.ModelStructName}} {
	repo := &{{.ModelStructName}}{}
	repo.initialize(ctx)
	return repo
}

`
