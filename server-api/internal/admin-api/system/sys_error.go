package system

import (
	systemDto "megin/internal/system/dto"
	"megin/internal/system/model"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysError @Tag 错误日志管理
type SysError struct{}

// @Summary 创建错误日志
// @Description 创建错误日志（无需认证）
func (this *SysError) CreateSysError(ctx *api.Context, req *model.SysError) (*api.Result[any], error) {
	err := systemService.NewSysError(ctx).CreateSysError(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除错误日志
// @Description 根据ID删除错误日志
func (this *SysError) DeleteSysError(ctx *api.Context, req *systemDto.DeleteSysErrorReq) (*api.Result[any], error) {
	err := systemService.NewSysError(ctx).DeleteSysError(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 批量删除错误日志
// @Description 批量删除错误日志
func (this *SysError) DeleteSysErrorByIds(ctx *api.Context, req *systemDto.DeleteSysErrorsReq) (*api.Result[any], error) {
	err := systemService.NewSysError(ctx).DeleteSysErrorByIds(req.IDs)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新错误日志
// @Description 更新错误日志信息
func (this *SysError) UpdateSysError(ctx *api.Context, req *systemDto.UpdateSysErrorReq) (*api.Result[any], error) {
	err := systemService.NewSysError(ctx).UpdateSysError(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 根据ID查询错误日志
// @Description 根据ID查询错误日志详情
func (this *SysError) FindSysError(ctx *api.Context, req *systemDto.GetSysErrorByIdReq) (*api.Result[systemDto.SysError], error) {
	result, err := systemService.NewSysError(ctx).GetSysError(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 获取错误日志列表
// @Description 分页获取错误日志列表
func (this *SysError) GetSysErrorList(ctx *api.Context, req *systemDto.SysErrorSearchReq) (*api.Result[systemDto.PageResult[systemDto.SysError]], error) {
	result, err := systemService.NewSysError(ctx).GetSysErrorInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 触发错误日志处理
// @Description 标记错误日志为处理中，1分钟后自动更新为处理完成
func (this *SysError) GetSysErrorSolution(ctx *api.Context, req *systemDto.GetSysErrorSolutionReq) (*api.Result[any], error) {
	id := uint(0)
	if req.ID != "" {
		var val uint64
		for _, c := range req.ID {
			val = val*10 + uint64(c-'0')
			id = uint(val)
		}
	}
	err := systemService.NewSysError(ctx).GetSysErrorSolution(id)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}
