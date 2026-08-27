package model

import (
	"time"

	"megin/internal/base"
)

const TableNameSysOperationRecord = "sys_operation_records"

type SysOperationRecord struct {
	base.SystemModel
	Ip           string        `gorm:"column:ip;comment:请求ip" json:"ip"`
	Method       string        `gorm:"column:method;comment:请求方法" json:"method"`
	Path         string        `gorm:"column:path;comment:请求路径" json:"path"`
	Status       int           `gorm:"column:status;comment:请求状态" json:"status"`
	Latency      time.Duration `gorm:"column:latency;comment:延迟" json:"latency"`
	Agent        string        `gorm:"type:text;column:agent;comment:代理" json:"agent"`
	ErrorMessage string        `gorm:"column:error_message;comment:错误信息" json:"errorMessage"`
	Body         string        `gorm:"type:text;column:body;comment:请求Body" json:"body"`
	Resp         string        `gorm:"type:text;column:resp;comment:响应Body" json:"resp"`
	UserID       int           `gorm:"column:user_id;comment:用户id" json:"userId"`
	User         SysUser       `json:"user"`
}

func (SysOperationRecord) TableName() string { return TableNameSysOperationRecord }
func (m SysOperationRecord) IsNil() bool     { return m.ID == 0 }
func (m SysOperationRecord) GetID() any      { return m.ID }
