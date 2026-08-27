package model

const TableNameSysUserAuthority = "sys_user_authority"

type SysUserAuthority struct {
	SysUserId                uint `gorm:"column:sys_user_id" json:"sysUserId"`
	SysAuthorityAuthorityId  uint `gorm:"column:sys_authority_authority_id" json:"sysAuthorityAuthorityId"`
}

func (SysUserAuthority) TableName() string { return TableNameSysUserAuthority }
func (m SysUserAuthority) IsNil() bool     { return m.SysUserId == 0 && m.SysAuthorityAuthorityId == 0 }
func (m SysUserAuthority) GetID() any      { return nil }