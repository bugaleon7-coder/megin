package dto

import (
	"time"
)

// SysDictionary 字典DTO
type SysDictionary struct {
	ID                   uint                  `json:"ID"`
	Name                 string                `json:"name"`
	Type                 string                `json:"type"`
	Status               *bool                 `json:"status"`
	Desc                 string                `json:"desc"`
	ParentID             *uint                 `json:"parentID"`
	CreatedAt            *time.Time            `json:"CreatedAt"`
	UpdatedAt            *time.Time            `json:"UpdatedAt"`
	Children             []SysDictionary       `json:"children"`
	SysDictionaryDetails []SysDictionaryDetail `json:"sysDictionaryDetails"`
}

// CreateDictionaryReq 创建字典请求
type CreateDictionaryReq struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Status   *bool  `json:"status"`
	Desc     string `json:"desc"`
	ParentID *uint  `json:"parentID"`
}

// UpdateDictionaryReq 更新字典请求
type UpdateDictionaryReq struct {
	ID       uint   `json:"id" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Status   *bool  `json:"status"`
	Desc     string `json:"desc"`
	ParentID *uint  `json:"parentID"`
}

// DictionarySearchReq 字典查询请求
type DictionarySearchReq struct {
	PageQuery
	Name string `json:"name" form:"name" binding:"omitempty"`
	Type string `json:"type" form:"type" binding:"omitempty"`
}

// DictionaryListReq 原版字典树列表仅支持按名称或类型模糊查询，不要求分页参数。
type DictionaryListReq struct {
	Name string `json:"name" form:"name" binding:"omitempty"`
}

type FindDictionaryReq struct {
	ID     uint   `json:"ID" form:"ID"`
	Type   string `json:"type" form:"type"`
	Status *bool  `json:"status" form:"status"`
}

type DictionaryResponse struct {
	Dictionary SysDictionary `json:"resysDictionary"`
}

// GetDictionaryByIdReq 根据ID查询字典请求
type GetDictionaryByIdReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// GetDictionaryByTypeReq 根据类型查询字典请求
type GetDictionaryByTypeReq struct {
	Type string `json:"type" form:"type" binding:"required"`
}

// ImportDictionaryReq 导入字典请求
type ImportDictionaryReq struct {
	Json string `json:"json" binding:"required"`
}
