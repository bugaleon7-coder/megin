package dto

// SystemConfigReq 系统配置更新请求
type SystemConfigReq struct {
	Config map[string]any `json:"config" binding:"required"`
}

// SystemConfigResponse 系统配置响应
type SystemConfigResponse struct {
	Config map[string]any `json:"config"`
}

// ServerInfoResponse 服务器信息响应
type ServerInfoResponse struct {
	Server map[string]any `json:"server"`
}
