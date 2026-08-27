package dto

import "megin/internal/system/model"

// CreateApiTokenReq 签发API Token请求
type CreateApiTokenReq struct {
	UserID      uint   `json:"userId" binding:"required"`
	AuthorityID uint   `json:"authorityId" binding:"required"`
	Days        int    `json:"days" binding:"required"`
	Remark      string `json:"remark"`
}

// GetApiTokenListReq API Token列表请求
type GetApiTokenListReq struct {
	PageQuery
	UserID uint  `json:"userId" form:"userId" binding:"omitempty"`
	Status *bool `json:"status" form:"status"`
}

// DeleteApiTokenReq 作废API Token请求
type DeleteApiTokenReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// ApiTokenResponse API Token列表响应
type ApiTokenResponse struct {
	List []model.SysApiToken `json:"list"`
}
