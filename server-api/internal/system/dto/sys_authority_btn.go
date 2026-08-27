package dto

// SysAuthorityBtnReq 角色按钮权限请求
type SysAuthorityBtnReq struct {
	AuthorityId   uint   `json:"authorityId" binding:"required"`
	SysMenuID     uint   `json:"sysMenuID" binding:"required"`
	Selected      []uint `json:"selected" binding:"required"`
}

// AuthorityBtnResponse 角色按钮权限响应
type AuthorityBtnResponse struct {
	Selected []uint `json:"selected"`
}

// CasbinInfo Casbin策略信息
type CasbinInfo struct {
	Path   string `json:"path"`
	Method string `json:"method"`
}

// CasbinCreateReq 创建Casbin策略请求
type CasbinCreateReq struct {
	AuthorityId uint         `json:"authorityId" binding:"required"`
	CasbinInfos []CasbinInfo `json:"casbinInfos" binding:"required"`
}

// PolicyPathResponse 策略路径响应
type PolicyPathResponse struct {
	Paths []CasbinInfo `json:"paths"`
}