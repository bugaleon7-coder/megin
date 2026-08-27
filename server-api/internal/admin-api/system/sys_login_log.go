package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysLoginLog @Tag 登录日志管理
type SysLoginLog struct{}

// @Summary 获取登录日志列表
// @Description 分页查询登录日志
func (this *SysLoginLog) GetSysLoginLogInfoList(ctx *api.Context, req *systemDto.LoginLogSearchReq) (*api.Result[systemDto.PageResult[systemDto.SysLoginLog]], error) {
	result, err := systemService.NewSysLoginLog(ctx).GetSysLoginLogInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 删除登录日志
// @Description 根据ID删除登录日志
func (this *SysLoginLog) DeleteSysLoginLog(ctx *api.Context, req *systemDto.DeleteLoginLogReq) (*api.Result[any], error) {
	err := systemService.NewSysLoginLog(ctx).DeleteSysLoginLog(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 批量删除登录日志
// @Description 批量删除登录日志
func (this *SysLoginLog) DeleteSysLoginLogs(ctx *api.Context, req *DeleteIdsReq) (*api.Result[any], error) {
	err := systemService.NewSysLoginLog(ctx).DeleteSysLoginLogs(req.Ids)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取登录日志详情
// @Description 根据ID获取登录日志详情
func (this *SysLoginLog) FindSysLoginLog(ctx *api.Context, req *systemDto.FindLoginLogReq) (*api.Result[systemDto.SysLoginLog], error) {
	result, err := systemService.NewSysLoginLog(ctx).FindSysLoginLog(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}
