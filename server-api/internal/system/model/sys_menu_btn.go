package model

import "megin/internal/base"

const TableNameSysBaseMenuBtn = "sys_base_menu_btns"

type SysBaseMenuBtn struct {
	base.SystemModel
	Name          string `gorm:"column:name;comment:按钮关键key" json:"name"`
	Desc          string `gorm:"column:desc;comment:按钮备注" json:"desc"`
	SysBaseMenuID uint   `gorm:"column:sys_base_menu_id;comment:菜单ID" json:"sysBaseMenuID"`
}

func (SysBaseMenuBtn) TableName() string { return TableNameSysBaseMenuBtn }
func (m SysBaseMenuBtn) IsNil() bool     { return m.ID == 0 }
func (m SysBaseMenuBtn) GetID() any      { return m.ID }
