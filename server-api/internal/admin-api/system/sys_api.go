package system

import (
	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysApi @Tag API管理
type SysApi struct{}

// @Summary 创建API
// @Description 创建系统API
func (this *SysApi) CreateApi(ctx *api.Context, req *systemDto.CreateApiReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).CreateApi(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除API
// @Description 删除系统API
func (this *SysApi) DeleteApi(ctx *api.Context, req *systemDto.DeleteApiReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).DeleteApi(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新API
// @Description 更新系统API信息
func (this *SysApi) UpdateApi(ctx *api.Context, req *systemDto.UpdateApiReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).UpdateApi(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取API详情
// @Description 根据ID获取API详细信息
func (this *SysApi) GetApiById(ctx *api.Context, req *systemDto.GetApiByIdReq) (*api.Result[systemDto.ApiResponse], error) {
	result, err := systemService.NewSysApi(ctx).GetApiById(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.ApiResponse{Api: *result})
}

// @Summary 获取API列表
// @Description 分页查询API列表
func (this *SysApi) GetApiList(ctx *api.Context, req *systemDto.GetApiListReq) (*api.Result[systemDto.PageResult[systemDto.SysApi]], error) {
	result, err := systemService.NewSysApi(ctx).GetApiList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 获取所有API
// @Description 获取所有API列表
func (this *SysApi) GetAllApis(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[systemDto.AllApisResponse], error) {
	result, err := systemService.NewSysApi(ctx).GetAllApis()
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.AllApisResponse{Apis: result})
}

// @Summary 获取API分组
// @Description 获取所有API分组列表
func (this *SysApi) GetApiGroups(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[systemDto.ApiGroupResponse], error) {
	result, err := systemService.NewSysApi(ctx).GetApiGroups()
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 批量删除API
// @Description 批量删除系统API
func (this *SysApi) DeleteApisByIds(ctx *api.Context, req *systemDto.DeleteApisByIdsReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).DeleteApisByIds(req.Ids)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 刷新Casbin缓存
// @Description 从数据库重新加载Casbin策略
func (this *SysApi) FreshCasbin(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).FreshCasbin()
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 同步API
// @Description 同步路由表中的API到数据库
func (this *SysApi) SyncApi(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[systemDto.SyncApiResponse], error) {
	result, err := systemService.NewSysApi(ctx).SyncApi()
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 忽略API
// @Description 设置API为忽略状态
func (this *SysApi) IgnoreApi(ctx *api.Context, req *systemDto.IgnoreApiReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).IgnoreApi(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 确认同步API
// @Description 确认执行API同步操作
func (this *SysApi) EnterSyncApi(ctx *api.Context, req *systemDto.EnterSyncApiReq) (*api.Result[any], error) {
	err := systemService.NewSysApi(ctx).EnterSyncApi(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取API角色列表
// @Description 根据路径和方法获取拥有该API权限的角色ID列表
func (this *SysApi) GetApiRoles(ctx *api.Context, req *systemDto.GetApiRolesReq) (*api.Result[[]uint], error) {
	casbinService := systemService.NewSysCasbin(ctx)
	authorityIds, err := casbinService.GetAuthoritiesByApi(req.Path, req.Method)
	if err != nil {
		return nil, err
	}
	if authorityIds == nil {
		authorityIds = []uint{}
	}
	return api.ResultData(authorityIds)
}

// @Summary 设置API角色列表
// @Description 全量覆盖某API关联的角色列表
func (this *SysApi) SetApiRoles(ctx *api.Context, req *systemDto.SetApiRolesReq) (*api.Result[any], error) {
	casbinService := systemService.NewSysCasbin(ctx)
	err := casbinService.SetApiAuthorities(req.Path, req.Method, req.AuthorityIds)
	if err != nil {
		return nil, err
	}
	// Refresh casbin cache to make policies take effect immediately
	_ = systemService.NewSysApi(ctx).FreshCasbin()
	return api.ResultSuccess()
}
