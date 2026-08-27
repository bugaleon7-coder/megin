package model

import "megin/internal/base"

const TableNameSysDictionary = "sys_dictionaries"

type SysDictionary struct {
	base.SystemModel
	Name                 string                `gorm:"column:name;comment:字典名（中）" json:"name"`
	Type                 string                `gorm:"column:type;comment:字典名（英）" json:"type"`
	Status               *bool                 `gorm:"column:status;comment:状态" json:"status"`
	Desc                 string                `gorm:"column:desc;comment:描述" json:"desc"`
	ParentID             *uint                 `gorm:"column:parent_id;comment:父级字典ID" json:"parentID"`
	Children             []SysDictionary       `gorm:"foreignKey:ParentID" json:"children"`
	SysDictionaryDetails []SysDictionaryDetail `gorm:"foreignKey:SysDictionaryID;references:ID" json:"sysDictionaryDetails"`
}

func (SysDictionary) TableName() string { return TableNameSysDictionary }
func (m SysDictionary) IsNil() bool     { return m.ID == 0 }
func (m SysDictionary) GetID() any      { return m.ID }
