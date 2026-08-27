package model

import "megin/internal/base"

const TableNameSysParams = "sys_params"

type SysParams struct {
	base.SystemModel
	Name  string `gorm:"column:name;comment:参数名称" json:"name"`
	Key   string `gorm:"column:key;comment:参数键" json:"key"`
	Value string `gorm:"column:value;comment:参数值" json:"value"`
	Desc  string `gorm:"column:desc;comment:参数说明" json:"desc"`
}

func (SysParams) TableName() string { return TableNameSysParams }
func (m SysParams) IsNil() bool     { return m.ID == 0 }
func (m SysParams) GetID() any      { return m.ID }
