package system

import (
	commonDto "megin/internal/module/common/dto"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysMenu @Tag 菜单管理
type SysMenu struct{}

// @Summary 获取菜单树
// @Description 获取全部菜单树形结构
func (this *SysMenu) GetMenuTree(ctx *api.Context, req *commonDto.BaseQueryByIdReq) (*api.Result[[]systemDto.SysMenu], error) {
	result, err := systemService.NewSysMenu(ctx).GetMenuTree()
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 获取基础菜单树
// @Description 获取基础菜单树形结构
func (this *SysMenu) GetBaseMenuTree(ctx *api.Context, req *dtoBaseReq) (*api.Result[systemDto.BaseMenuTreeResponse], error) {
	result, err := systemService.NewSysMenu(ctx).GetBaseMenuTree()
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.BaseMenuTreeResponse{Menus: result})
}

// @Summary 添加菜单
// @Description 添加基础菜单
func (this *SysMenu) AddBaseMenu(ctx *api.Context, req *systemDto.AddBaseMenuReq) (*api.Result[any], error) {
	err := systemService.NewSysMenu(ctx).AddBaseMenu(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除菜单
// @Description 删除基础菜单
func (this *SysMenu) DeleteBaseMenu(ctx *api.Context, req *systemDto.DeleteMenuReq) (*api.Result[any], error) {
	err := systemService.NewSysMenu(ctx).DeleteBaseMenu(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新菜单
// @Description 更新基础菜单信息
func (this *SysMenu) UpdateBaseMenu(ctx *api.Context, req *systemDto.UpdateBaseMenuReq) (*api.Result[any], error) {
	err := systemService.NewSysMenu(ctx).UpdateBaseMenu(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取菜单详情
// @Description 根据ID获取菜单详细信息
func (this *SysMenu) GetBaseMenuById(ctx *api.Context, req *systemDto.GetMenuByIdReq) (*api.Result[systemDto.BaseMenuResponse], error) {
	result, err := systemService.NewSysMenu(ctx).GetBaseMenuById(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.BaseMenuResponse{Menu: *result})
}

// @Summary 获取菜单权限
// @Description 根据角色ID获取已授权的菜单
func (this *SysMenu) GetMenuAuthority(ctx *api.Context, req *systemDto.GetMenuAuthorityReq) (*api.Result[systemDto.MenuAuthorityResponse], error) {
	result, err := systemService.NewSysMenu(ctx).GetMenuAuthority(req.AuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.MenuAuthorityResponse{Menus: result})
}

// @Summary 添加菜单权限
// @Description 为角色添加菜单权限
func (this *SysMenu) AddMenuAuthority(ctx *api.Context, req *systemDto.AddMenuAuthorityReq) (*api.Result[any], error) {
	err := systemService.NewSysMenu(ctx).AddMenuAuthority(uint(ctx.AdminInfo.RoleId), req.AuthorityId, extractMenuIds(req.Menus))
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取菜单列表
// @Description 获取基础菜单列表
func (this *SysMenu) GetMenuInfoList(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[[]systemDto.SysBaseMenu], error) {
	result, err := systemService.NewSysMenu(ctx).GetMenuInfoList()
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 获取菜单树
// @Description 获取当前角色的菜单树（前端用）
func (this *SysMenu) GetMenu(ctx *api.Context, req *dtoBaseReq) (*api.Result[systemDto.MenuTreeResponse], error) {
	authorityId := uint(ctx.AdminInfo.RoleId)
	menus, err := systemService.NewSysMenu(ctx).GetMenuTreeByAuthority(authorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.MenuTreeResponse{Menus: menus})
}

// GetMenuRoles 获取菜单角色列表
// @Summary 获取拥有指定菜单的角色ID列表
// @Description 根据菜单ID获取关联的角色ID列表
func (this *SysMenu) GetMenuRoles(ctx *api.Context, req *systemDto.GetMenuRolesReq) (*api.Result[systemDto.MenuRolesResponse], error) {
	authorityIds, err := systemService.NewSysMenu(ctx).GetAuthoritiesByMenuId(req.MenuId)
	if err != nil {
		return nil, err
	}
	defaultRouterAuthorityIds, err := systemService.NewSysMenu(ctx).GetDefaultRouterAuthorityIds(req.MenuId)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.MenuRolesResponse{
		AuthorityIds:              authorityIds,
		DefaultRouterAuthorityIds: defaultRouterAuthorityIds,
	})
}

// SetMenuRoles 设置菜单角色列表
// @Summary 全量覆盖某菜单关联的角色列表
// @Description 设置某菜单关联的角色ID列表
func (this *SysMenu) SetMenuRoles(ctx *api.Context, req *systemDto.SetMenuRolesReq) (*api.Result[any], error) {
	err := systemService.NewSysMenu(ctx).SetMenuAuthorities(req.MenuId, req.AuthorityIds)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func extractMenuIds(menus []systemDto.SysBaseMenu) []uint {
	ids := make([]uint, 0, len(menus))
	seen := make(map[uint]struct{})
	var collect func([]systemDto.SysBaseMenu)
	collect = func(items []systemDto.SysBaseMenu) {
		for _, menu := range items {
			if _, ok := seen[menu.ID]; !ok {
				seen[menu.ID] = struct{}{}
				ids = append(ids, menu.ID)
			}
			collect(menu.Children)
		}
	}
	collect(menus)
	return ids
}

// dtoBaseReq 空请求
type dtoBaseReq struct{}
