package api

import (
	"context"
	"fmt"
	bizcache "megin/internal/cache"
	commonDto "megin/internal/module/common/dto"
	"megin/pkg/logger"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Context struct {
	TraceId     string
	ProjectPath string
	AdminInfo   *commonDto.Claims // 管理后台管理员JWT信息，仅 admin-api 认证成功后有值
	UserInfo    *commonDto.Claims // C端用户JWT信息，仅 api 认证成功后有值
	Token       string            // 当前请求携带的原始JWT Token
	GinCtx      *gin.Context      //gin的Context,只有在gin接口调用时有值
	error       error
	Log         *logger.Logger //zap log

	Tx *gorm.DB //事务，如果有的话
}

func (ctx *Context) EnableTx(tx *gorm.DB) *Context {
	ctx.Tx = tx
	return ctx
}

func (ctx *Context) Param(key string) string {
	return ctx.GinCtx.Param(key)
}

// RequestKeyPrefix 根据当前请求生成统一的缓存 key 前缀。
// 生成规则：
// - category: 业务类别，例如 lock、limit
// - method: HTTP 方法，例如 post
// - route: 当前路由模板，例如 admin-api:article:update
func (ctx *Context) RequestKeyPrefix(category string) string {
	method := "unknown"
	route := "unknown"
	if ctx != nil && ctx.GinCtx != nil {
		if ctx.GinCtx.Request != nil && ctx.GinCtx.Request.Method != "" {
			method = strings.ToLower(ctx.GinCtx.Request.Method)
		}
		route = ctx.GinCtx.FullPath()
		if route == "" && ctx.GinCtx.Request != nil && ctx.GinCtx.Request.URL != nil {
			route = ctx.GinCtx.Request.URL.Path
		}
	}
	route = strings.Trim(route, "/")
	if route == "" {
		route = "unknown"
	}
	route = strings.NewReplacer("/", ":", ":", "", "*", "").Replace(route)
	return fmt.Sprintf("%s:%s:%s", category, method, route)
}

// RequestContext 返回当前请求绑定的标准库 context。
// 当 Gin 上下文不存在时，返回后台 context，保证非 HTTP 场景也能安全调用缓存能力。
func (ctx *Context) RequestContext() context.Context {
	if ctx != nil && ctx.GinCtx != nil && ctx.GinCtx.Request != nil {
		return ctx.GinCtx.Request.Context()
	}
	return context.Background()
}

// Redis 返回当前请求绑定的 Redis 访问器。
// 该访问器会自动复用请求 context，减少业务层重复传参。
func (ctx *Context) Redis() *RedisAccessor {
	return NewRedisAccessor(ctx.RequestContext(), bizcache.Redis())
}

// LocalCache 返回默认本地缓存访问器。
// 该访问器适合普通本地缓存场景，底层复用 internal/cache 默认缓存池。
func (ctx *Context) LocalCache() *LocalCacheAccessor {
	return NewLocalCacheAccessor(ctx.RequestContext(), bizcache.Local())
}

// Locker 返回当前请求绑定的分布式锁访问器。
// 当前实现基于 Redis SetNX，并通过 token 校验释放锁，避免误删他人锁。
func (ctx *Context) Locker() *LockerAccessor {
	return NewLockerAccessor(ctx.RequestContext(), bizcache.Redis(), ctx.RequestKeyPrefix("lock"))
}

// Limiter 返回当前请求绑定的限流访问器。
// 当前实现基于 Redis SetNX + TTL，适合“同一个 key 在固定时间窗内只允许一次”的场景。
func (ctx *Context) Limiter() *LimiterAccessor {
	return NewLimiterAccessor(ctx.RequestContext(), bizcache.Redis(), ctx.RequestKeyPrefix("limit"))
}

type HandleFunc func(ctx *Context)

// 每个请求一个Context
func NewContext(ginCtx *gin.Context) (*Context, error) {
	ctx := new(Context)
	ctx.GinCtx = ginCtx
	ctx.Log = logger.New()
	if claims, ok := getClaims(ginCtx, commonDto.AdminApiJwtClaims); ok {
		ctx.AdminInfo = claims
		ctx.Token = getTokenValue(ginCtx, commonDto.AdminApiClaimToken)
	} else if claims, ok := getClaims(ginCtx, commonDto.ApiJwtClaims); ok {
		ctx.UserInfo = claims
		ctx.Token = getTokenValue(ginCtx, commonDto.ApiClaimToken)
	}
	return ctx, nil
}

func getClaims(ginCtx *gin.Context, key string) (*commonDto.Claims, bool) {
	jwtClaims, ok := ginCtx.Get(key)
	if !ok {
		return nil, false
	}
	claims, ok := jwtClaims.(*commonDto.Claims)
	return claims, ok
}

func getTokenValue(ginCtx *gin.Context, key string) string {
	tokenValue, ok := ginCtx.Get(key)
	if !ok {
		return ""
	}
	token, ok := tokenValue.(string)
	if !ok {
		return ""
	}
	return token
}
