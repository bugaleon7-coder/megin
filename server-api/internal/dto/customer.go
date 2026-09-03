package dto

import (
	"time"
)

type Customer struct {
	ID                 uint       `json:"ID" form:"ID"`        // 客户 ID。
	CreatedAt          *time.Time `json:"CreatedAt,omitempty"` // 创建时间。
	UpdatedAt          *time.Time `json:"UpdatedAt,omitempty"` // 最后更新时间。
	CustomerName       string     `json:"customerName"`        // 客户名称。
	CustomerPhoneData  string     `json:"customerPhoneData"`   // 客户联系电话。
	SysUserID          uint       `json:"sysUserId"`           // 所属系统用户 ID。
	SysUserAuthorityID uint       `json:"sysUserAuthorityID"`  // 所属系统用户的角色 ID。
	SysUser            any        `json:"sysUser"`             // 关联的系统用户信息。
}
type CreateCustomerReq struct {
	CustomerName      string `json:"customerName" binding:"required"`      // 客户名称，必填。
	CustomerPhoneData string `json:"customerPhoneData" binding:"required"` // 客户联系电话，必填。
}
type UpdateCustomerReq struct {
	ID                uint   `json:"ID" binding:"required"`                // 要更新的客户 ID。
	CustomerName      string `json:"customerName" binding:"required"`      // 更新后的客户名称。
	CustomerPhoneData string `json:"customerPhoneData" binding:"required"` // 更新后的客户联系电话。
}
type DeleteCustomerReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"` // 要删除的客户 ID。
}
type GetCustomerReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"` // 要查询的客户 ID。
}
type GetCustomerListReq struct {
	PageNo   int `form:"page" json:"page" binding:"required,min=1"`         // 页码，从 1 开始。
	PageSize int `form:"pageSize" json:"pageSize" binding:"required,min=1"` // 每页记录数。
}
type CustomerResponse struct {
	Customer Customer `json:"customer"` // 客户详情。
}
type CustomerPageResult[T any] struct {
	PageNo    int   `json:"page"`     // 当前页码。
	PageSize  int   `json:"pageSize"` // 每页记录数。
	TotalSize int64 `json:"total"`    // 总记录数。
	List      []T   `json:"list"`     // 当前页客户列表。
}
