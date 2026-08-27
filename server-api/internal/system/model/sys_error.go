package model

import "megin/internal/base"

const TableNameSysError = "sys_error"

type SysError struct {
	base.SystemModel
	Form     *string `gorm:"column:form;comment:错误来源;type:text" json:"form"`
	Info     *string `gorm:"column:info;comment:错误内容;type:text" json:"info"`
	Level    string  `gorm:"column:level;comment:日志等级" json:"level"`
	Solution *string `gorm:"column:solution;comment:解决方案;type:text" json:"solution"`
	Status   string  `gorm:"column:status;comment:处理状态;type:varchar(20);default:未处理" json:"status"`
}

func (SysError) TableName() string { return TableNameSysError }
func (m SysError) IsNil() bool     { return m.ID == 0 }
func (m SysError) GetID() any      { return m.ID }
