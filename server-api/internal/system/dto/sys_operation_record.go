package dto

import (
	"time"
)

// SysOperationRecord 操作记录DTO
type SysOperationRecord struct {
	ID           uint       `json:"ID"`
	Ip           string     `json:"ip"`
	Method       string     `json:"method"`
	Path         string     `json:"path"`
	Status       int        `json:"status"`
	Latency      int64      `json:"latency"`
	Agent        string     `json:"agent"`
	ErrorMessage string     `json:"errorMessage"`
	Body         string     `json:"body"`
	Resp         string     `json:"resp"`
	UserID       int        `json:"userId"`
	CreatedAt    *time.Time `json:"CreatedAt"`
	User         SysUser    `json:"user"`
}

// OperationRecordSearchReq 操作记录查询请求
type OperationRecordSearchReq struct {
	PageQuery
	Method string `json:"method" form:"method" binding:"omitempty"`
	Path   string `json:"path" form:"path" binding:"omitempty"`
	Status int    `json:"status" form:"status" binding:"omitempty"`
}
