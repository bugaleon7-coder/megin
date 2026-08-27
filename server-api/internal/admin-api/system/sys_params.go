package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysParams @Tag 系统参数管理
type SysParams struct{}

// @Summary 创建参数
// @Description 创建系统参数
func (this *SysParams) CreateSysParams(ctx *api.Context, req *systemDto.CreateParamsReq) (*api.Result[any], error) {
	err := systemService.NewSysParams(ctx).CreateSysParams(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除参数
// @Description 删除系统参数
func (this *SysParams) DeleteSysParams(ctx *api.Context, req *systemDto.DeleteParamsReq) (*api.Result[any], error) {
	err := systemService.NewSysParams(ctx).DeleteSysParams(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新参数
// @Description 更新系统参数信息
func (this *SysParams) UpdateSysParams(ctx *api.Context, req *systemDto.UpdateParamsReq) (*api.Result[any], error) {
	err := systemService.NewSysParams(ctx).UpdateSysParams(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取参数详情
// @Description 根据ID获取系统参数
func (this *SysParams) GetSysParams(ctx *api.Context, req *systemDto.GetParamsByIdReq) (*api.Result[systemDto.SysParams], error) {
	result, err := systemService.NewSysParams(ctx).GetSysParams(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 根据Key获取参数
// @Description 根据Key获取系统参数
func (this *SysParams) GetSysParamsByKey(ctx *api.Context, req *GetSysParamsByKeyReq) (*api.Result[systemDto.SysParams], error) {
	result, err := systemService.NewSysParams(ctx).GetSysParamsByKey(req.Key)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 获取参数列表
// @Description 分页查询系统参数列表
func (this *SysParams) GetSysParamsInfoList(ctx *api.Context, req *systemDto.ParamsSearchReq) (*api.Result[systemDto.PageResult[systemDto.SysParams]], error) {
	result, err := systemService.NewSysParams(ctx).GetSysParamsInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 批量删除参数
// @Description 批量删除系统参数
func (this *SysParams) DeleteSysParamsByIds(ctx *api.Context, req *DeleteIdsReq) (*api.Result[any], error) {
	err := systemService.NewSysParams(ctx).DeleteSysParamsByIds(req.Ids)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// GetSysParamsByKeyReq 根据Key查询参数请求
type GetSysParamsByKeyReq struct {
	Key string `json:"key" form:"key" binding:"required"`
}
