package model

import (
	"time"

	"megin/internal/base"
)

const TableNameSysApiToken = "sys_api_tokens"

type SysApiToken struct {
	base.SystemModel
	UserID      uint      `gorm:"column:user_id;comment:用户ID" json:"userId"`
	User        SysUser   `gorm:"foreignKey:UserID;references:ID" json:"user"`
	AuthorityID uint      `gorm:"column:authority_id;comment:角色ID" json:"authorityId"`
	Token       string    `gorm:"column:token;type:text;comment:Token" json:"token"`
	Status      bool      `gorm:"column:status;default:true;comment:状态" json:"status"`
	ExpiresAt   time.Time `gorm:"column:expires_at;comment:过期时间" json:"expiresAt"`
	Remark      string    `gorm:"column:remark;comment:备注" json:"remark"`
}

func (SysApiToken) TableName() string { return TableNameSysApiToken }
func (m SysApiToken) IsNil() bool     { return m.ID == 0 }
func (m SysApiToken) GetID() any      { return m.ID }
