package model

import "megin/internal/base"

const TableNameSysJwtBlacklist = "jwt_blacklists"

type JwtBlacklist struct {
	base.SystemModel
	Jwt string `gorm:"type:text;column:jwt;comment:jwt" json:"jwt"`
}

func (JwtBlacklist) TableName() string { return TableNameSysJwtBlacklist }
func (m JwtBlacklist) IsNil() bool     { return m.ID == 0 }
func (m JwtBlacklist) GetID() any      { return m.ID }
