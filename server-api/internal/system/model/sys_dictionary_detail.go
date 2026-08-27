package model

import "megin/internal/base"

const TableNameSysDictionaryDetail = "sys_dictionary_details"

type SysDictionaryDetail struct {
	base.SystemModel
	Label           string                `gorm:"column:label;comment:展示值" json:"label"`
	Value           string                `gorm:"column:value;comment:字典值" json:"value"`
	Extend          string                `gorm:"column:extend;comment:扩展值" json:"extend"`
	Status          *bool                 `gorm:"column:status;comment:启用状态" json:"status"`
	Sort            int                   `gorm:"column:sort;comment:排序标记" json:"sort"`
	SysDictionaryID int                   `gorm:"column:sys_dictionary_id;comment:关联标记" json:"sysDictionaryID"`
	ParentID        *uint                 `gorm:"column:parent_id;comment:父级字典详情ID" json:"parentID"`
	Level           int                   `gorm:"column:level;comment:层级深度" json:"level"`
	Path            string                `gorm:"column:path;comment:层级路径" json:"path"`
	Children        []SysDictionaryDetail `gorm:"foreignKey:ParentID" json:"children"`
}

func (SysDictionaryDetail) TableName() string { return TableNameSysDictionaryDetail }
func (m SysDictionaryDetail) IsNil() bool     { return m.ID == 0 }
func (m SysDictionaryDetail) GetID() any      { return m.ID }
