package model

import (
	"time"

	"gorm.io/gorm"
)

const TableNameSysAuthority = "sys_authorities"

type SysAuthority struct {
	CreatedAt       *time.Time      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt       *time.Time      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt       gorm.DeletedAt  `gorm:"column:deleted_at;index" json:"-"`
	AuthorityId     uint            `gorm:"column:authority_id;not null;unique;primaryKey;comment:角色ID;size:90" json:"authorityId"`
	AuthorityName   string          `gorm:"column:authority_name;comment:角色名" json:"authorityName"`
	ParentId        *uint           `gorm:"column:parent_id;comment:父角色ID" json:"parentId"`
	DefaultRouter   string          `gorm:"column:default_router;comment:默认菜单;default:dashboard" json:"defaultRouter"`
	DataAuthorityId []*SysAuthority `gorm:"many2many:sys_data_authority_id" json:"dataAuthorityId"`
	Children        []SysAuthority  `gorm:"-" json:"children"`
	SysBaseMenus    []SysBaseMenu   `gorm:"many2many:sys_authority_menus" json:"menus"`
}

func (SysAuthority) TableName() string { return TableNameSysAuthority }
func (m SysAuthority) IsNil() bool     { return m.AuthorityId == 0 }
func (m SysAuthority) GetID() any      { return m.AuthorityId }
