package model

import "megin/internal/base"

const (
	TableNameSysApi       = "sys_apis"
	TableNameSysIgnoreApi = "sys_ignore_apis"
)

type SysApi struct {
	base.SystemModel
	Path        string `gorm:"column:path;comment:api路径" json:"path"`
	Description string `gorm:"column:description;comment:api中文描述" json:"description"`
	ApiGroup    string `gorm:"column:api_group;comment:api组" json:"apiGroup"`
	Method      string `gorm:"column:method;default:POST;comment:方法" json:"method"`
}

type SysIgnoreApi struct {
	ID     uint   `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Path   string `gorm:"column:path;" json:"path"`
	Method string `gorm:"column:method;default:POST" json:"method"`
	Flag   bool   `gorm:"-" json:"flag"`
}

func (SysApi) TableName() string       { return TableNameSysApi }
func (m SysApi) IsNil() bool           { return m.ID == 0 }
func (m SysApi) GetID() any            { return m.ID }
func (SysIgnoreApi) TableName() string { return TableNameSysIgnoreApi }
func (m SysIgnoreApi) IsNil() bool     { return m.ID == 0 }
func (m SysIgnoreApi) GetID() any      { return m.ID }
