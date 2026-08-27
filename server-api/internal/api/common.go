package api

import (
	commonDto "megin/internal/module/common/dto"
	contextApi "megin/pkg/context/api"
)

// Common @Tag 前台公共接口
// Common 收口无需登录即可访问的前台公共能力。
type Common struct{}

// Health @Summary 前台 API 健康检查
// @Description 用于确认前台 API 服务是否正常；该接口无需 Token，并作为指定接口 IP 限流的默认示例。
// @Description 请求字段：无。
// @Description 返回字段：data 固定为 ok，表示 API 服务能够正常处理请求。
func (h *Common) Health(ctx *contextApi.Context, req *commonDto.EmptyReq) (*contextApi.Result[string], error) {
	return contextApi.ResultData("ok")
}
