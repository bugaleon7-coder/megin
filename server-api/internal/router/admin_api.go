package router

import (
	handler "megin/internal/admin-api"
	sysApi "megin/internal/admin-api/system"
	"megin/internal/middleware"
	sysRouter "megin/internal/system/router"
	"megin/pkg/context/router"
)

func NewAdminProtectedGroup(routeRegistry *router.RouteRegistry) *router.RouteGroup {
	adminApiGroup := routeRegistry.Group("admin-api")
	adminApiGroup.Use(middleware.AdminApiAuthTokenRequired())
	return adminApiGroup
}

func NewAdminPublicGroup(routeRegistry *router.RouteRegistry) *router.RouteGroup {
	return routeRegistry.Group("admin-api")
}

// 管理后台API
func AdminApiRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {

	// article routes
	article := &handler.Article{}
	router.GET(adminApiGroup, "/article/detail", article.Detail)
	router.POST(adminApiGroup, "/article/create", article.Create)
	router.POST(adminApiGroup, "/article/update", article.Update)
	router.POST(adminApiGroup, "/article/delete", article.Delete)
	router.GET(adminApiGroup, "/article/pageList", article.PageList)

	customer := &handler.Customer{}
	router.POST(adminApiGroup, "/customer/customer", customer.Create)
	router.PUT(adminApiGroup, "/customer/customer", customer.Update)
	router.DELETEWithBody(adminApiGroup, "/customer/customer", customer.Delete)
	router.GET(adminApiGroup, "/customer/customer", customer.Detail)
	router.GET(adminApiGroup, "/customer/customerList", customer.List)

	// 前台 API 限流规则管理。
	rateLimit := &handler.RateLimitRule{}
	router.POST(adminApiGroup, "/rate-limit/create", rateLimit.Create)
	router.PUT(adminApiGroup, "/rate-limit/update", rateLimit.Update)
	router.PUT(adminApiGroup, "/rate-limit/changeStatus", rateLimit.ChangeStatus)
	router.DELETE(adminApiGroup, "/rate-limit/delete", rateLimit.Delete)
	router.GET(adminApiGroup, "/rate-limit/detail", rateLimit.Detail)
	router.GET(adminApiGroup, "/rate-limit/pageList", rateLimit.PageList)
	router.POST(adminApiGroup, "/rate-limit/refresh", rateLimit.Refresh)

	return adminApiGroup
}

// 系统管理模块 - 管理后台路由
func InitSystemAdminRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	sysRouter.SysUserRouter(adminApiGroup)
	sysRouter.SysAuthorityRouter(adminApiGroup)
	sysRouter.SysMenuRouter(adminApiGroup)
	sysRouter.SysApiRouter(adminApiGroup)
	sysRouter.SysCasbinRouter(adminApiGroup)
	sysRouter.SysDictionaryRouter(adminApiGroup)
	sysRouter.SysDictionaryDetailRouter(adminApiGroup)
	sysRouter.SysOperationRecordRouter(adminApiGroup)
	sysRouter.SysLoginLogRouter(adminApiGroup)
	sysRouter.SysParamsRouter(adminApiGroup)
	sysRouter.SysApiTokenRouter(adminApiGroup)
	sysRouter.SysVersionRouter(adminApiGroup)
	sysRouter.SysSystemRouter(adminApiGroup)
	sysRouter.SysJwtRouter(adminApiGroup)
	sysRouter.SysAuthorityBtnRouter(adminApiGroup)
	sysRouter.SysErrorRouter(adminApiGroup)
	return adminApiGroup
}

// 系统管理模块 - 无需认证的公开路由（admin-api）
func InitSystemApiRouter(noAuthGroup *router.RouteGroup) *router.RouteGroup {
	sysUser := &sysApi.SysUser{}
	router.POST(noAuthGroup, "/user/login", sysUser.Login)
	router.POST(noAuthGroup, "/user/captcha", sysUser.Captcha)
	router.GET(noAuthGroup, "/user/loginConfig", sysUser.LoginConfig)
	sysRouter.SysErrorPublicRouter(noAuthGroup)

	return noAuthGroup
}
