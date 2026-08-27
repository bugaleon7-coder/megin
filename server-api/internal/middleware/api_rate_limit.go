package middleware

import (
	"math"
	"megin/internal/config"
	commonDto "megin/internal/module/common/dto"
	rateLimitRuntime "megin/internal/module/rate_limit/runtime"
	"megin/pkg/context/api"
	"megin/pkg/errs"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ApiIPRateLimit 创建前台 API 的 IP 限流中间件。
//
// 中间件挂载在 /api 父路由组，因此公开接口和鉴权接口都会执行。运行时先使用
// HTTP Method + Gin FullPath 查找接口 IP 规则；接口规则不存在时才回退全局 IP 规则，
// 两条规则不会叠加消费。
//
// ctx.ClientIP() 会按照 Gin 的可信代理配置解析请求来源。生产环境必须正确配置可信代理，
// 避免所有客户端被识别为同一个代理 IP，或者客户端通过伪造转发请求头绕过限制。
func ApiIPRateLimit() gin.HandlerFunc {
	if !config.GetConfig().APIRateLimit.Enable {
		return passThrough()
	}
	manager := rateLimitRuntime.DefaultManager()
	if manager == nil {
		panic("API限流管理器未初始化")
	}
	return func(ctx *gin.Context) {
		decision := manager.AllowIP(ctx.Request.Method, routePath(ctx), ctx.ClientIP())
		if !decision.Allowed {
			writeRateLimitResponse(ctx, decision)
			return
		}
		ctx.Next()
	}
}

// ApiUIDRateLimit 创建前台 API 的 UID 限流中间件。
//
// 中间件必须位于 ApiAuthTokenRequired 之后。鉴权中间件已经验证 Token 并把 Claims
// 写入 Gin Context，本中间件只读取 UserID。运行时优先使用接口 UID 规则，没有时才
// 回退全局 UID 规则；同一用户签发多个 Token 仍然共享同一个 UID 令牌桶。
func ApiUIDRateLimit() gin.HandlerFunc {
	conf := config.GetConfig().APIRateLimit
	if !conf.Enable {
		return passThrough()
	}
	manager := rateLimitRuntime.DefaultManager()
	if manager == nil {
		panic("API限流管理器未初始化")
	}
	return func(ctx *gin.Context) {
		value, exists := ctx.Get(commonDto.ApiJwtClaims)
		claims, ok := value.(*commonDto.Claims)
		if !exists || !ok || claims.UserID <= 0 {
			ctx.JSON(http.StatusOK, api.Failed[any](errs.NewBusinessError(403, "token已失效,请重新登录")))
			ctx.Abort()
			return
		}
		decision := manager.AllowUID(ctx.Request.Method, routePath(ctx), claims.UserID)
		if !decision.Allowed {
			writeRateLimitResponse(ctx, decision)
			return
		}
		ctx.Next()
	}
}

func routePath(ctx *gin.Context) string {
	path := ctx.FullPath()
	if path == "" && ctx.Request != nil && ctx.Request.URL != nil {
		path = ctx.Request.URL.Path
	}
	return path
}

// writeRateLimitResponse 输出统一限流响应，并给客户端提供建议重试秒数。
func writeRateLimitResponse(ctx *gin.Context, decision rateLimitRuntime.Decision) {
	retrySeconds := int(math.Ceil(decision.RetryAfter.Seconds()))
	if retrySeconds < 1 {
		retrySeconds = 1
	}
	ctx.Header("Retry-After", strconv.Itoa(retrySeconds))
	ctx.Set("rate_limit_rule_id", decision.RuleID)
	ctx.JSON(http.StatusOK, api.Failed[any](errs.NewBusinessError(429, "请求过于频繁，请稍后再试")))
	ctx.Abort()
}

// passThrough 返回不改变请求链路的中间件，用于总开关关闭时保持路由组装一致。
func passThrough() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}
