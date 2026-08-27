package model

const TableNameSysAuthorityBtn = "sys_authority_btns"

type SysAuthorityBtn struct {
	AuthorityId      uint           `gorm:"column:authority_id;comment:角色ID" json:"authorityId"`
	SysMenuID        uint           `gorm:"column:sys_menu_id;comment:菜单ID" json:"sysMenuID"`
	SysBaseMenuBtnID uint           `gorm:"column:sys_base_menu_btn_id;comment:菜单按钮ID" json:"sysBaseMenuBtnID"`
	SysBaseMenuBtn   SysBaseMenuBtn `gorm:"foreignKey:SysBaseMenuBtnID;references:ID;comment:按钮详情" json:"sysBaseMenuBtn"`
}

func (SysAuthorityBtn) TableName() string { return TableNameSysAuthorityBtn }
func (m SysAuthorityBtn) IsNil() bool     { return m.AuthorityId == 0 && m.SysMenuID == 0 && m.SysBaseMenuBtnID == 0 }
func (m SysAuthorityBtn) GetID() any      { return nil }