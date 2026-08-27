package dto

import (
	"time"
)

// SysParams 系统参数DTO
type SysParams struct {
	ID        uint       `json:"ID"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	Value     string     `json:"value"`
	Desc      string     `json:"desc"`
	CreatedAt *time.Time `json:"CreatedAt"`
	UpdatedAt *time.Time `json:"UpdatedAt"`
}

// CreateParamsReq 创建参数请求
type CreateParamsReq struct {
	Name  string `json:"name" binding:"required"`
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	Desc  string `json:"desc"`
}

// UpdateParamsReq 更新参数请求
type UpdateParamsReq struct {
	ID    uint   `json:"ID" binding:"required"`
	Name  string `json:"name" binding:"required"`
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
	Desc  string `json:"desc"`
}

// ParamsSearchReq 参数查询请求
type ParamsSearchReq struct {
	PageQuery
	Name string `json:"name" form:"name" binding:"omitempty"`
	Key  string `json:"key" form:"key" binding:"omitempty"`
}

// GetParamsByIdReq 根据ID查询参数请求
type GetParamsByIdReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}
