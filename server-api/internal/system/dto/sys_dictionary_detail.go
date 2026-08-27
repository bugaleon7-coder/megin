package dto

import (
	"time"
)

// SysDictionaryDetail 字典详情DTO
type SysDictionaryDetail struct {
	ID              uint                  `json:"ID"`
	Label           string                `json:"label"`
	Value           string                `json:"value"`
	Extend          string                `json:"extend"`
	Status          *bool                 `json:"status"`
	Disabled        bool                  `json:"disabled"`
	Sort            int                   `json:"sort"`
	SysDictionaryID int                   `json:"sysDictionaryID"`
	ParentID        *uint                 `json:"parentID"`
	Level           int                   `json:"level"`
	Path            string                `json:"path"`
	CreatedAt       *time.Time            `json:"CreatedAt"`
	UpdatedAt       *time.Time            `json:"UpdatedAt"`
	Children        []SysDictionaryDetail `json:"children"`
}

// CreateDictionaryDetailReq 创建字典详情请求
type CreateDictionaryDetailReq struct {
	Label           string `json:"label" binding:"required"`
	Value           string `json:"value" binding:"required"`
	Extend          string `json:"extend"`
	Status          *bool  `json:"status"`
	Sort            int    `json:"sort"`
	SysDictionaryID int    `json:"sysDictionaryID" binding:"required"`
	ParentID        *uint  `json:"parentID"`
}

// UpdateDictionaryDetailReq 更新字典详情请求
type UpdateDictionaryDetailReq struct {
	ID              uint   `json:"ID" binding:"required"`
	Label           string `json:"label" binding:"required"`
	Value           string `json:"value" binding:"required"`
	Extend          string `json:"extend"`
	Status          *bool  `json:"status"`
	Sort            int    `json:"sort"`
	SysDictionaryID int    `json:"sysDictionaryID" binding:"required"`
	ParentID        *uint  `json:"parentID"`
}

// DictionaryDetailSearchReq 字典详情查询请求
type DictionaryDetailSearchReq struct {
	PageQuery
	Label           string `json:"label" form:"label" binding:"omitempty"`
	Value           string `json:"value" form:"value" binding:"omitempty"`
	SysDictionaryID int    `json:"sysDictionaryID" form:"sysDictionaryID" binding:"omitempty"`
	ParentID        *uint  `json:"parentID" form:"parentID" binding:"omitempty"`
	Level           *int   `json:"level" form:"level" binding:"omitempty"`
}

// GetDictionaryDetailByIdReq 根据ID查询字典详情请求
type GetDictionaryDetailByIdReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// GetDetailsByParentReq 根据父ID查询字典详情请求
type GetDetailsByParentReq struct {
	SysDictionaryID int   `json:"sysDictionaryID" form:"sysDictionaryID" binding:"required"`
	ParentID        *uint `json:"parentID" form:"parentID"`
	IncludeChildren bool  `json:"includeChildren" form:"includeChildren"`
}

// GetDictionaryTreeReq 根据字典ID获取树形结构请求
type GetDictionaryTreeReq struct {
	SysDictionaryID int `json:"sysDictionaryID" form:"sysDictionaryID" binding:"required"`
}

// GetDictionaryTreeByTypeReq 根据字典类型获取树形结构请求
type GetDictionaryTreeByTypeReq struct {
	Type string `json:"type" form:"type" binding:"required"`
}

// GetDictionaryPathReq 获取字典详情路径请求
type GetDictionaryPathReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// DictionaryDetailResponse 字典详情响应
type DictionaryDetailResponse struct {
	DictionaryDetail SysDictionaryDetail `json:"reSysDictionaryDetail"`
}

// DictionaryDetailListResponse 树形/列表响应
type DictionaryDetailListResponse struct {
	List []SysDictionaryDetail `json:"list"`
}
