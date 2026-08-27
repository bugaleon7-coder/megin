package dto

import "time"

// Meta 菜单Meta信息
type Meta struct {
	ActiveName     string `json:"activeName"`
	KeepAlive      bool   `json:"keepAlive"`
	DefaultMenu    bool   `json:"defaultMenu"`
	Title          string `json:"title" binding:"required"`
	Icon           string `json:"icon"`
	CloseTab       bool   `json:"closeTab"`
	TransitionType string `json:"transitionType"`
}

// SysBaseMenu 基础菜单DTO
type SysBaseMenu struct {
	ID         uint               `json:"ID"`
	CreatedAt  *time.Time         `json:"CreatedAt,omitempty"`
	UpdatedAt  *time.Time         `json:"UpdatedAt,omitempty"`
	MenuLevel  uint               `json:"-"`
	ParentId   uint               `json:"parentId"`
	Path       string             `json:"path"`
	Name       string             `json:"name"`
	Hidden     bool               `json:"hidden"`
	Component  string             `json:"component"`
	Sort       int                `json:"sort"`
	Meta       Meta               `json:"meta"`
	Children   []SysBaseMenu      `json:"children"`
	Parameters []SysBaseMenuParam `json:"parameters"`
	MenuBtn    []SysBaseMenuBtn   `json:"menuBtn"`
}

// SysBaseMenuParam 菜单参数DTO
type SysBaseMenuParam struct {
	SysBaseMenuID uint   `json:"sysBaseMenuID"`
	Type          string `json:"type"`
	Key           string `json:"key"`
	Value         string `json:"value"`
}

// SysBaseMenuBtn 菜单按钮DTO
type SysBaseMenuBtn struct {
	ID   uint   `json:"ID"`
	Name string `json:"name"`
	Desc string `json:"desc"`
}

// AddBaseMenuReq 添加菜单请求
type AddBaseMenuReq struct {
	ParentId  uint   `json:"parentId" binding:"omitempty"`
	Path      string `json:"path" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Hidden    bool   `json:"hidden"`
	Component string `json:"component" binding:"required"`
	Sort      int    `json:"sort" binding:"omitempty"`
	Meta      Meta   `json:"meta" binding:"required"`
	Title     string `json:"title"`
	Icon      string `json:"icon" binding:"omitempty"`
}

// UpdateBaseMenuReq 更新菜单请求
type UpdateBaseMenuReq struct {
	ID        uint   `json:"id" binding:"required"`
	ParentId  uint   `json:"parentId" binding:"omitempty"`
	Path      string `json:"path" binding:"required"`
	Name      string `json:"name" binding:"required"`
	Hidden    bool   `json:"hidden"`
	Component string `json:"component" binding:"required"`
	Sort      int    `json:"sort" binding:"omitempty"`
	Meta      Meta   `json:"meta" binding:"required"`
	Title     string `json:"title"`
	Icon      string `json:"icon" binding:"omitempty"`
}

// GetMenuByIdReq 根据ID查询菜单请求
type GetMenuByIdReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// AddMenuAuthorityReq 添加菜单权限请求
type AddMenuAuthorityReq struct {
	Menus       []SysBaseMenu `json:"menus" binding:"required"`
	AuthorityId uint          `json:"authorityId" binding:"required"`
}

// GetMenuAuthorityReq 查询菜单权限请求
type GetMenuAuthorityReq struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"`
}

// SysMenu 菜单（带权限）DTO
type SysMenu struct {
	SysBaseMenu
	MenuId     uint               `json:"menuId"`
	Children   []SysMenu          `json:"children"`
	Parameters []SysBaseMenuParam `json:"parameters"`
	Btns       map[string]uint    `json:"btns"`
}

// MenuTreeResponse 菜单树响应
type MenuTreeResponse struct {
	Menus []SysMenu `json:"menus"`
}

// BaseMenuTreeResponse 基础菜单树响应
type BaseMenuTreeResponse struct {
	Menus []SysBaseMenu `json:"menus"`
}

// BaseMenuResponse 保持 GVA 前端 getBaseMenuById 所依赖的 data.menu 响应结构。
type BaseMenuResponse struct {
	Menu SysBaseMenu `json:"menu"`
}

// MenuAuthorityResponse 菜单权限响应
type MenuAuthorityResponse struct {
	Menus []SysBaseMenu `json:"menus"`
}

// GetMenuRolesReq 获取菜单角色列表请求
type GetMenuRolesReq struct {
	MenuId uint `json:"menuId" form:"menuId" binding:"required"`
}

// SetMenuRolesReq 设置菜单角色列表请求
type SetMenuRolesReq struct {
	MenuId       uint   `json:"menuId" binding:"required"`
	AuthorityIds []uint `json:"authorityIds" binding:"required"`
}

// MenuRolesResponse 菜单角色列表响应
type MenuRolesResponse struct {
	AuthorityIds              []uint `json:"authorityIds"`
	DefaultRouterAuthorityIds []uint `json:"defaultRouterAuthorityIds"`
}
