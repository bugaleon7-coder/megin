package dto

import (
	"time"

	"megin/internal/system/model"
)

// SysVersion 版本管理列表/详情响应
type SysVersion struct {
	ID          uint       `json:"ID"`
	VersionName *string    `json:"versionName"`
	VersionCode *string    `json:"versionCode"`
	Description *string    `json:"description"`
	VersionData *string    `json:"versionData"`
	CreatedAt   *time.Time `json:"CreatedAt"`
	UpdatedAt   *time.Time `json:"UpdatedAt"`
}

// SysVersionSearch 分页查询请求
type SysVersionSearch struct {
	PageQuery
	CreatedAtRange []time.Time `json:"createdAtRange" form:"createdAtRange[]"`
	VersionName    *string     `json:"versionName" form:"versionName"`
	VersionCode    *string     `json:"versionCode" form:"versionCode"`
}

// GetSysVersionReq 根据ID查询版本请求
type GetSysVersionReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// DeleteSysVersionByIdsReq 批量删除版本请求
type DeleteSysVersionByIdsReq struct {
	IDs []uint `json:"IDs" form:"IDs[]" binding:"required,min=1"`
}

// ExportVersionRequest 导出版本请求
type ExportVersionRequest struct {
	VersionName string `json:"versionName" binding:"required"`
	VersionCode string `json:"versionCode" binding:"required"`
	Description string `json:"description"`
	MenuIds     []uint `json:"menuIds"`
	ApiIds      []uint `json:"apiIds"`
	DictIds     []uint `json:"dictIds"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	ExportTime  string `json:"exportTime"`
}

// ImportVersionRequest 导入版本请求
type ImportVersionRequest struct {
	VersionInfo      VersionInfo            `json:"version" binding:"required"`
	ExportMenu       []model.SysBaseMenu    `json:"menus"`
	ExportApi        []model.SysApi         `json:"apis"`
	ExportDictionary []model.SysDictionary  `json:"dictionaries"`
}

// ExportVersionResponse 导出版本响应
type ExportVersionResponse struct {
	Version      VersionInfo             `json:"version"`
	Menus        []model.SysBaseMenu     `json:"menus"`
	Apis         []model.SysApi          `json:"apis"`
	Dictionaries []model.SysDictionary   `json:"dictionaries"`
}
