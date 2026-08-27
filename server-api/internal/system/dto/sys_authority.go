package dto

import (
	"time"
)

// SysAuthority 角色DTO
type SysAuthority struct {
	CreatedAt       *time.Time      `json:"CreatedAt"`
	UpdatedAt       *time.Time      `json:"UpdatedAt"`
	AuthorityId     uint            `json:"authorityId"`
	AuthorityName   string          `json:"authorityName"`
	ParentId        *uint           `json:"parentId"`
	DefaultRouter   string          `json:"defaultRouter"`
	DataAuthorityId []*SysAuthority `json:"dataAuthorityId"`
	Children        []SysAuthority  `json:"children"`
}

// CreateAuthorityReq 创建角色请求
type CreateAuthorityReq struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"`
	AuthorityName string `json:"authorityName" binding:"required"`
	ParentId      *uint  `json:"parentId" binding:"omitempty"`
	DefaultRouter string `json:"defaultRouter" binding:"omitempty"`
}

// UpdateAuthorityReq 更新角色请求
type UpdateAuthorityReq struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"`
	AuthorityName string `json:"authorityName" binding:"required"`
	ParentId      *uint  `json:"parentId" binding:"omitempty"`
	DefaultRouter string `json:"defaultRouter" binding:"omitempty"`
}

// CopyAuthorityReq 复制角色请求
type CopyAuthorityReq struct {
	Authority      CreateAuthorityReq `json:"authority" binding:"required"`
	OldAuthorityId uint               `json:"oldAuthorityId" binding:"required"`
}

// AuthorityIdReq 角色ID请求
type AuthorityIdReq struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"`
}

// SetDataAuthorityReq 设置数据权限请求
type SetDataAuthorityReq struct {
	AuthorityId     uint            `json:"authorityId" binding:"required"`
	DataAuthorityId []*SysAuthority `json:"dataAuthorityId" binding:"required"`
}

// SetMenuAuthorityReq 设置菜单权限请求
type SetMenuAuthorityReq struct {
	AuthorityId uint   `json:"authorityId" binding:"required"`
	MenuIds     []uint `json:"menuIds" binding:"required"`
}

// SetRoleUsersReq 设置角色用户请求
type SetRoleUsersReq struct {
	AuthorityId uint   `json:"authorityId" binding:"required"`
	UserIds     []uint `json:"userIds" binding:"required"`
}

// GetAuthorityListReq 角色列表查询请求（不要求分页参数）
type GetAuthorityListReq struct {
}

// AuthorityResponse 角色响应
type AuthorityResponse struct {
	Authority SysAuthority `json:"authority"`
}

// AuthorityCopyResponse 角色复制响应
type AuthorityCopyResponse struct {
	Authority      SysAuthority `json:"authority"`
	OldAuthorityId uint         `json:"oldAuthorityId"`
}
