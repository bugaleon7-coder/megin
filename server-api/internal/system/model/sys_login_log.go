package model

import "megin/internal/base"

const TableNameSysLoginLog = "sys_login_logs"

type SysLoginLog struct {
	base.SystemModel
	Username     string `gorm:"column:username;comment:用户名" json:"username"`
	Ip           string `gorm:"column:ip;comment:请求ip" json:"ip"`
	Status       bool   `gorm:"column:status;comment:登录状态" json:"status"`
	ErrorMessage string `gorm:"column:error_message;comment:错误信息" json:"errorMessage"`
	Agent        string `gorm:"column:agent;comment:代理" json:"agent"`
	UserID       uint   `gorm:"column:user_id;comment:用户id" json:"userId"`
}

func (SysLoginLog) TableName() string { return TableNameSysLoginLog }
func (m SysLoginLog) IsNil() bool     { return m.ID == 0 }
func (m SysLoginLog) GetID() any      { return m.ID }
