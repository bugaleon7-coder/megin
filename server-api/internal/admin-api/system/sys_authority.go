package system

import (
	systemDto "megin/internal/system/dto"
	systemModel "megin/internal/system/model"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysAuthority @Tag 角色管理
type SysAuthority struct{}

// @Summary 创建角色
// @Description 创建系统角色
func (this *SysAuthority) CreateAuthority(ctx *api.Context, req *systemDto.CreateAuthorityReq) (*api.Result[any], error) {
	_, err := systemService.NewSysAuthority(ctx).CreateAuthority(uint(ctx.AdminInfo.RoleId), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 复制角色
// @Description 从已有角色复制创建新角色
func (this *SysAuthority) CopyAuthority(ctx *api.Context, req *systemDto.CopyAuthorityReq) (*api.Result[any], error) {
	_, err := systemService.NewSysAuthority(ctx).CopyAuthority(uint(ctx.AdminInfo.RoleId), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 更新角色
// @Description 更新角色信息
func (this *SysAuthority) UpdateAuthority(ctx *api.Context, req *systemDto.UpdateAuthorityReq) (*api.Result[any], error) {
	err := systemService.NewSysAuthority(ctx).UpdateAuthority(uint(ctx.AdminInfo.RoleId), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除角色
// @Description 删除系统角色
func (this *SysAuthority) DeleteAuthority(ctx *api.Context, req *systemDto.DeleteAuthorityReq) (*api.Result[any], error) {
	err := systemService.NewSysAuthority(ctx).DeleteAuthority(uint(ctx.AdminInfo.RoleId), req.AuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取角色列表
// @Description 获取角色树列表（非分页）
func (this *SysAuthority) GetAuthorityList(ctx *api.Context, req *systemDto.GetAuthorityListReq) (*api.Result[[]systemDto.SysAuthority], error) {
	// Get admin authority ID from JWT context
	adminAuthorityId := uint(ctx.AdminInfo.RoleId)
	result, err := systemService.NewSysAuthority(ctx).GetAuthorityInfoList(adminAuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// @Summary 获取角色信息
// @Description 根据角色ID获取详细信息
func (this *SysAuthority) GetAuthorityInfo(ctx *api.Context, req *systemDto.DeleteAuthorityReq) (*api.Result[systemDto.SysAuthority], error) {
	result, err := systemService.NewSysAuthority(ctx).GetAuthorityInfo(req.AuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 设置数据权限
// @Description 设置角色的数据权限
func (this *SysAuthority) SetDataAuthority(ctx *api.Context, req *systemModel.SysAuthority) (*api.Result[any], error) {
	err := systemService.NewSysAuthority(ctx).SetDataAuthority(uint(ctx.AdminInfo.RoleId), *req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 设置菜单权限
// @Description 设置角色的菜单权限
func (this *SysAuthority) SetMenuAuthority(ctx *api.Context, req *systemDto.SetMenuAuthorityReq) (*api.Result[any], error) {
	err := systemService.NewSysAuthority(ctx).SetMenuAuthority(uint(ctx.AdminInfo.RoleId), req.AuthorityId, req.MenuIds)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取角色用户ID列表
// @Description 获取属于该角色的所有用户ID
func (this *SysAuthority) GetUserIdsByAuthorityId(ctx *api.Context, req *systemDto.DeleteAuthorityReq) (*api.Result[[]uint], error) {
	ids, err := systemService.NewSysAuthority(ctx).GetUserIdsByAuthorityId(uint(ctx.AdminInfo.RoleId), req.AuthorityId)
	if err != nil {
		return nil, err
	}
	if ids == nil {
		ids = []uint{}
	}
	return api.ResultData(ids)
}

// @Summary 设置角色用户
// @Description 为角色设置用户
func (this *SysAuthority) SetRoleUsers(ctx *api.Context, req *systemDto.SetRoleUsersReq) (*api.Result[any], error) {
	err := systemService.NewSysAuthority(ctx).SetRoleUsers(uint(ctx.AdminInfo.RoleId), req.AuthorityId, req.UserIds)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}
