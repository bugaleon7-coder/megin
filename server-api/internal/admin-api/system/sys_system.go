package system

import (
	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysSystem @Tag 系统配置管理
type SysSystem struct{}

// @Summary 获取配置文件内容
// @Description 获取系统配置文件内容
func (this *SysSystem) GetSystemConfig(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[systemDto.SystemConfigResponse], error) {
	result, err := systemService.NewSysSystem(ctx).GetSystemConfig()
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 设置配置文件内容
// @Description 设置系统配置文件内容
func (this *SysSystem) SetSystemConfig(ctx *api.Context, req *systemDto.SystemConfigReq) (*api.Result[any], error) {
	if err := systemService.NewSysSystem(ctx).SetSystemConfig(req); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取服务器信息
// @Description 获取服务器运行状态信息
func (this *SysSystem) GetServerInfo(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[systemDto.ServerInfoResponse], error) {
	server, err := systemService.NewSysSystem(ctx).GetServerInfo()
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.ServerInfoResponse{Server: server})
}

// @Summary 重载系统
// @Description 重载系统配置
func (this *SysSystem) ReloadSystem(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[any], error) {
	if err := systemService.NewSysSystem(ctx).ReloadSystem(); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}
