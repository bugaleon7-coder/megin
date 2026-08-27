package dto

import (
	"time"
)

// SysError 错误日志DTO
type SysError struct {
	ID        uint       `json:"ID"`
	Form      *string    `json:"form"`
	Info      *string    `json:"info"`
	Level     string     `json:"level"`
	Solution  *string    `json:"solution"`
	Status    string     `json:"status"`
	CreatedAt *time.Time `json:"CreatedAt"`
	UpdatedAt *time.Time `json:"UpdatedAt"`
}

// CreateSysErrorReq 创建错误日志请求
type CreateSysErrorReq struct {
	Form  *string `json:"form"`
	Info  *string `json:"info"`
	Level string  `json:"level"`
}

// UpdateSysErrorReq 更新错误日志请求
type UpdateSysErrorReq struct {
	ID       uint    `json:"id" binding:"required"`
	Solution *string `json:"solution"`
	Status   string  `json:"status"`
}

// SysErrorSearchReq 错误日志查询请求
type SysErrorSearchReq struct {
	PageQuery
	CreatedAtRange []time.Time `json:"createdAtRange" form:"createdAtRange[]"`
	Form           *string     `json:"form" form:"form"`
	Info           *string     `json:"info" form:"info"`
}

// GetSysErrorByIdReq 根据ID查询错误日志请求
type GetSysErrorByIdReq struct {
	ID uint `json:"id" form:"ID" binding:"required"`
}

// GetSysErrorSolutionReq 触发错误处理请求
type GetSysErrorSolutionReq struct {
	ID string `json:"id" form:"id" binding:"required"`
}

// DeleteSysErrorReq 删除错误日志请求
type DeleteSysErrorReq struct {
	ID uint `json:"id" form:"ID" binding:"required"`
}

// DeleteSysErrorsReq 批量删除错误日志请求
type DeleteSysErrorsReq struct {
	IDs []string `json:"ids" form:"IDs[]" binding:"required"`
}
