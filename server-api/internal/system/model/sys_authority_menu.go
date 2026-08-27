package model

const TableNameSysAuthorityMenu = "sys_authority_menus"

type SysAuthorityMenu struct {
	MenuId      string `gorm:"column:sys_base_menu_id;comment:菜单ID" json:"menuId"`
	AuthorityId string `gorm:"column:sys_authority_authority_id;comment:角色ID" json:"-"`
}

func (SysAuthorityMenu) TableName() string { return TableNameSysAuthorityMenu }
func (m SysAuthorityMenu) IsNil() bool     { return m.MenuId == "" && m.AuthorityId == "" }
func (m SysAuthorityMenu) GetID() any      { return nil }