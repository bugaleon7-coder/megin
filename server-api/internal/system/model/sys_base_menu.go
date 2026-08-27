package model

import "megin/internal/base"

const TableNameSysBaseMenu = "sys_base_menus"

type SysBaseMenu struct {
	base.SystemModel
	MenuLevel     uint   `gorm:"column:menu_level;" json:"-"`
	ParentId      uint   `gorm:"column:parent_id;comment:父菜单ID" json:"parentId"`
	Path          string `gorm:"column:path;comment:路由path" json:"path"`
	Name          string `gorm:"column:name;comment:路由name" json:"name"`
	Hidden        bool   `gorm:"column:hidden;comment:是否在列表隐藏" json:"hidden"`
	Component     string `gorm:"column:component;comment:对应前端文件路径" json:"component"`
	Sort          int    `gorm:"column:sort;comment:排序标记" json:"sort"`
	Meta          `gorm:"embedded" json:"meta"`
	SysAuthoritys []SysAuthority         `gorm:"many2many:sys_authority_menus" json:"-"`
	Children      []SysBaseMenu          `gorm:"-" json:"children"`
	Parameters    []SysBaseMenuParameter `gorm:"foreignKey:SysBaseMenuID;references:ID" json:"-"`
	MenuBtn       []SysBaseMenuBtn       `gorm:"foreignKey:SysBaseMenuID;references:ID" json:"-"`
}

type Meta struct {
	ActiveName     string `json:"activeName" gorm:"column:active_name;"`
	KeepAlive      bool   `json:"keepAlive" gorm:"column:keep_alive;"`
	DefaultMenu    bool   `json:"defaultMenu" gorm:"column:default_menu;"`
	Title          string `json:"title" gorm:"column:title;"`
	Icon           string `json:"icon" gorm:"column:icon;"`
	CloseTab       bool   `json:"closeTab" gorm:"column:close_tab;"`
	TransitionType string `json:"transitionType" gorm:"column:transition_type;"`
}

type SysBaseMenuParameter struct {
	ID            uint   `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	SysBaseMenuID uint   `gorm:"column:sys_base_menu_id;" json:"sysBaseMenuID"`
	Type          string `gorm:"column:type;" json:"type"`
	Key           string `gorm:"column:key;" json:"key"`
	Value         string `gorm:"column:value;" json:"value"`
}

func (SysBaseMenu) TableName() string          { return TableNameSysBaseMenu }
func (m SysBaseMenu) IsNil() bool              { return m.ID == 0 }
func (m SysBaseMenu) GetID() any               { return m.ID }
func (SysBaseMenuParameter) TableName() string { return "sys_base_menu_parameters" }
func (m SysBaseMenuParameter) IsNil() bool     { return m.ID == 0 }
func (m SysBaseMenuParameter) GetID() any      { return m.ID }
