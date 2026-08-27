package dto

import (
	"time"
)

// SysApi API DTO
type SysApi struct {
	ID          uint       `json:"ID"`
	Path        string     `json:"path"`
	Description string     `json:"description"`
	ApiGroup    string     `json:"apiGroup"`
	Method      string     `json:"method"`
	CreatedAt   *time.Time `json:"CreatedAt"`
	UpdatedAt   *time.Time `json:"UpdatedAt"`
}

// ApiResponse 保持 GVA 前端 getApiById 所依赖的 data.api 响应结构。
type ApiResponse struct {
	Api SysApi `json:"api"`
}

// CreateApiReq 创建API请求
type CreateApiReq struct {
	Path        string `json:"path" binding:"required"`
	Description string `json:"description" binding:"required"`
	ApiGroup    string `json:"apiGroup" binding:"required"`
	Method      string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
}

// UpdateApiReq 更新API请求
type UpdateApiReq struct {
	ID          uint   `json:"id" binding:"required"`
	Path        string `json:"path" binding:"required"`
	Description string `json:"description" binding:"required"`
	ApiGroup    string `json:"apiGroup" binding:"required"`
	Method      string `json:"method" binding:"required,oneof=GET POST PUT DELETE PATCH"`
}

// GetApiByIdReq 根据ID查询API请求
type GetApiByIdReq struct {
	ID uint `json:"id" binding:"required"`
}

// GetApiListReq API列表查询请求
type GetApiListReq struct {
	PageQuery
	Path        string `json:"path" form:"path" binding:"omitempty"`
	Description string `json:"description" form:"description" binding:"omitempty"`
	ApiGroup    string `json:"apiGroup" form:"apiGroup" binding:"omitempty"`
	Method      string `json:"method" form:"method" binding:"omitempty"`
}

// DeleteApisByIdsReq 批量删除API请求
type DeleteApisByIdsReq struct {
	Ids []int `json:"ids" binding:"required,min=1"`
}

// ApiGroupResponse API分组响应
type ApiGroupResponse struct {
	Groups      []string          `json:"groups"`
	ApiGroupMap map[string]string `json:"apiGroupMap"`
}

// AllApisResponse 所有API响应
type AllApisResponse struct {
	Apis []SysApi `json:"apis"`
}

// SyncApiResponse 同步API响应
type SyncApiResponse struct {
	NewApis    []SysApi `json:"newApis"`
	DeleteApis []SysApi `json:"deleteApis"`
	IgnoreApis []SysApi `json:"ignoreApis"`
}

// IgnoreApiReq 忽略API请求
type IgnoreApiReq struct {
	Path   string `json:"path" binding:"required"`
	Method string `json:"method" binding:"required"`
	Flag   bool   `json:"flag"`
}

// EnterSyncApiReq 确认同步API请求
type EnterSyncApiReq struct {
	NewApis    []SysApiSyncItem `json:"newApis"`
	DeleteApis []SysApiSyncItem `json:"deleteApis"`
}

// SysApiSyncItem 同步API项
type SysApiSyncItem struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// GetApiRolesReq 获取API角色请求
type GetApiRolesReq struct {
	Path   string `json:"path" form:"path" binding:"required"`
	Method string `json:"method" form:"method" binding:"required"`
}

// SetApiRolesReq 设置API角色请求
type SetApiRolesReq struct {
	Path         string `json:"path" binding:"required"`
	Method       string `json:"method" binding:"required"`
	AuthorityIds []uint `json:"authorityIds" binding:"required"`
}
