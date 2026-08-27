package dto

import (
	"time"
)

// SysLoginLog 登录日志DTO
type SysLoginLog struct {
	ID           uint       `json:"ID"`
	Username     string     `json:"username"`
	Ip           string     `json:"ip"`
	Status       bool       `json:"status"`
	ErrorMessage string     `json:"errorMessage"`
	Agent        string     `json:"agent"`
	UserID       uint       `json:"userId"`
	CreatedAt    *time.Time `json:"CreatedAt"`
}

// FindLoginLogReq 查询登录日志请求
type FindLoginLogReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}
type LoginLogSearchReq struct {
	PageQuery
	Username string `json:"username" form:"username" binding:"omitempty"`
	Ip       string `json:"ip" form:"ip" binding:"omitempty"`
	Status   string `json:"status" form:"status" binding:"omitempty"`
}
