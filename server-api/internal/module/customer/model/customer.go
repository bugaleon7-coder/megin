package model

import (
	"megin/internal/base"
	systemModel "megin/internal/system/model"
)

const TableNameCustomer = "exa_customers"

type Customer struct {
	base.SystemModel
	CustomerName       string              `gorm:"column:customer_name;comment:客户名" json:"customerName"`
	CustomerPhoneData  string              `gorm:"column:customer_phone_data;comment:客户手机号" json:"customerPhoneData"`
	SysUserID          uint                `gorm:"column:sys_user_id;comment:管理ID" json:"sysUserId"`
	SysUserAuthorityID uint                `gorm:"column:sys_user_authority_id;comment:管理角色ID" json:"sysUserAuthorityID"`
	SysUser            systemModel.SysUser `gorm:"foreignKey:SysUserID;references:ID" json:"sysUser"`
}

func (Customer) TableName() string { return TableNameCustomer }
func (m Customer) IsNil() bool     { return m.ID == 0 }
func (m Customer) GetID() any      { return m.ID }
