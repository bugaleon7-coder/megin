package system

import (
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysApiToken @Tag API Token管理
type SysApiToken struct{}

// @Summary 签发API Token
// @Description 为用户签发API Token
func (this *SysApiToken) CreateApiToken(ctx *api.Context, req *systemDto.CreateApiTokenReq) (*api.Result[map[string]string], error) {
	token, err := systemService.NewSysApiToken(ctx).CreateApiToken(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(map[string]string{"token": token})
}

// @Summary 获取API Token列表
// @Description 分页查询API Token列表
func (this *SysApiToken) GetApiTokenList(ctx *api.Context, req *systemDto.GetApiTokenListReq) (*api.Result[systemDto.PageResult[model.SysApiToken]], error) {
	result, err := systemService.NewSysApiToken(ctx).GetApiTokenList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 作废API Token
// @Description 根据ID作废API Token
func (this *SysApiToken) DeleteApiToken(ctx *api.Context, req *systemDto.DeleteApiTokenReq) (*api.Result[any], error) {
	if err := systemService.NewSysApiToken(ctx).DeleteApiToken(req.ID); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}
