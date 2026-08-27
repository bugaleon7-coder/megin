package system

import (
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
	"strings"
)

// SysJwt @Tag JWT管理
type SysJwt struct{}

// @Summary 将JWT加入黑名单
// @Description 将JWT Token加入黑名单
func (this *SysJwt) JsonInBlacklist(ctx *api.Context, req *systemDto.JsonInBlacklistReq) (*api.Result[any], error) {
	token := blacklistToken(ctx, req)
	if token == "" {
		return nil, systemService.NewSysJwt(ctx).ErrorMessage("token不能为空")
	}

	err := systemService.NewSysJwt(ctx).JsonInBlacklist(token)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func blacklistToken(ctx *api.Context, req *systemDto.JsonInBlacklistReq) string {
	if req != nil && req.Token != "" {
		return req.Token
	}
	if ctx == nil || ctx.GinCtx == nil {
		return ""
	}

	authHeader := ctx.GinCtx.GetHeader("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return authHeader[7:]
	}

	for _, header := range []string{"X-Token", "x-token", "Token"} {
		if token := ctx.GinCtx.GetHeader(header); token != "" {
			return token
		}
	}

	token, _ := ctx.GinCtx.Cookie("x-token")
	return token
}
