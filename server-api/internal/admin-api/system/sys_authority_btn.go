package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
)

// SysAuthorityBtn @Tag 按钮权限管理
type SysAuthorityBtn struct{}

// @Summary 获取按钮权限
// @Description 根据角色ID和菜单ID获取按钮权限
func (this *SysAuthorityBtn) GetAuthorityBtn(ctx *api.Context, req *GetAuthorityBtnReq) (*api.Result[systemDto.AuthorityBtnResponse], error) {
	result, err := systemService.NewSysAuthorityBtn(ctx).GetAuthorityBtn(req.AuthorityId, req.MenuID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 设置按钮权限
// @Description 为角色设置菜单的按钮权限
func (this *SysAuthorityBtn) SetAuthorityBtn(ctx *api.Context, req *systemDto.SysAuthorityBtnReq) (*api.Result[any], error) {
	err := systemService.NewSysAuthorityBtn(ctx).SetAuthorityBtn(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// GetAuthorityBtnReq 查询按钮权限请求
type GetAuthorityBtnReq struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"`
	MenuID      uint `json:"menuID" form:"menuID" binding:"required"`
}
