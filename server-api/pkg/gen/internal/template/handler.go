package template

// Model used as a variable because it cannot load template file after packed, params still can pass file
const Handler = NotEditMark + `
package handler

import (
	"megin/internal/api/dto"
	"megin/internal/service"
	"megin/pkg/context/api"
	"megin/pkg/validate"
)

// {{.StructComment}}
type {{.ModelStructName}} struct {

}

//根据ID查询详情
func (this *{{.ModelStructName}}) Detail(ctx *api.Context) *api.Result {
	var req dto.BaseQueryByIdReq
	validate.BindWithPanic(ctx, &req)
	/*
	model, err := service.New{{.ModelStructName}}(ctx).GetById(req.ID)
	if err != nil {
		return api.Failed(err)
	}
	return api.Success(model)
	*/
	return api.Success()
}

`
