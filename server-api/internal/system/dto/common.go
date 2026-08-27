package dto

import commonDto "megin/internal/module/common/dto"

// JWT相关请求
type JsonInBlacklistReq struct {
	Token string `json:"token"`
}

// GetRedisJWTReq 获取Redis中JWT请求
type GetRedisJWTReq struct {
	UserID int `json:"userID" form:"userID" binding:"required"`
}

// DeleteOperationRecordReq 删除操作记录请求
type DeleteOperationRecordReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DeleteLoginLogReq 删除登录日志请求
type DeleteLoginLogReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// DeleteDictionaryReq 删除字典请求
type DeleteDictionaryReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DeleteDictionaryDetailReq 删除字典详情请求
type DeleteDictionaryDetailReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DeleteAuthorityReq 删除角色请求
type DeleteAuthorityReq struct {
	AuthorityId uint `json:"authorityId" form:"authorityId" binding:"required"`
}

// GetUserInfoReq 获取用户信息请求
type GetUserInfoReq struct {
	ID uint `json:"id" form:"id" binding:"omitempty"`
}

// DeleteApiReq 删除API请求
type DeleteApiReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DeleteMenuReq 删除菜单请求
type DeleteMenuReq struct {
	ID uint `json:"id" form:"id" binding:"required"`
}

// DeleteParamsReq 删除参数请求
type DeleteParamsReq struct {
	ID uint `json:"ID" form:"ID" binding:"required"`
}

// SyncApiReq 同步API请求
type SyncApiReq struct {
	Delete bool `json:"delete" form:"delete"`
}

// DeleteApisByIdsReqV2 批量删除API请求
type DeleteApisByIdsReqV2 struct {
	commonDto.BaseDeleteByIdReq
}

var DefaultCasbinInfos = []CasbinInfo{
	{Path: "/menu/getMenu", Method: "POST"},
	{Path: "/jwt/jsonInBlacklist", Method: "POST"},
	{Path: "/user/login", Method: "POST"},
	{Path: "/user/changePassword", Method: "POST"},
	{Path: "/user/setUserAuthority", Method: "POST"},
	{Path: "/user/getUserInfo", Method: "GET"},
	{Path: "/user/setSelfInfo", Method: "PUT"},
	{Path: "/user/setSelfSetting", Method: "PUT"},
	{Path: "/user/getTotpStatus", Method: "GET"},
	{Path: "/user/initTotp", Method: "POST"},
	{Path: "/user/enableTotp", Method: "POST"},
	{Path: "/user/disableTotp", Method: "POST"},
}
