package dto

import (
	"time"

	systemDto "megin/internal/system/dto"
	systemModel "megin/internal/system/model"
)

type Customer struct {
	ID                 uint                `json:"ID" form:"ID"`
	CreatedAt          *time.Time          `json:"CreatedAt,omitempty"`
	UpdatedAt          *time.Time          `json:"UpdatedAt,omitempty"`
	CustomerName       string              `json:"customerName"`
	CustomerPhoneData  string              `json:"customerPhoneData"`
	SysUserID          uint                `json:"sysUserId"`
	SysUserAuthorityID uint                `json:"sysUserAuthorityID"`
	SysUser            systemModel.SysUser `json:"sysUser"`
}

type CreateCustomerReq struct {
	CustomerName      string `json:"customerName" binding:"required"`
	CustomerPhoneData string `json:"customerPhoneData" binding:"required"`
}

type UpdateCustomerReq struct {
	ID                uint   `json:"ID" binding:"required"`
	CustomerName      string `json:"customerName" binding:"required"`
	CustomerPhoneData string `json:"customerPhoneData" binding:"required"`
}

type DeleteCustomerReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

type GetCustomerReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

type GetCustomerListReq struct {
	systemDto.PageQuery
}

type CustomerResponse struct {
	Customer Customer `json:"customer"`
}

type PageResult[T any] = systemDto.PageResult[T]
