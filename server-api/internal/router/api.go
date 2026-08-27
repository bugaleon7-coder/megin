package router

import (
	"megin/internal/api"
	"megin/internal/middleware"
	articleModule "megin/internal/module/article"
	"megin/pkg/context/router"
)

// 前端API
func InitApiRouter(routeRegistry *router.RouteRegistry) *router.RouteRegistry {
	// 所有前台接口先按客户端 IP 限流。
	rootGroup := routeRegistry.Group("api")
	rootGroup.Use(middleware.ApiIPRateLimit())

	// 需要登录的接口在鉴权成功后继续按 UID 限流。
	apiGroup := rootGroup.Group("")
	apiGroup.Use(middleware.ApiAuthTokenRequired())
	apiGroup.Use(middleware.ApiUIDRateLimit())

	// 无需登录的接口只按客户端 IP 限流。
	noAuthGroup := rootGroup.Group("")

	common := &api.Common{}
	router.GET(noAuthGroup, "/health", common.Health)

	user := &api.User{}
	router.POST(noAuthGroup, "/user/register", user.Register)
	router.POST(noAuthGroup, "/user/login", user.Login)
	router.GET(apiGroup, "/user/info", user.Info)

	article := &api.Article{}
	router.GET(noAuthGroup, "/article/detail", article.Detail, articleModule.DetailOptions...)
	router.POST(noAuthGroup, "/article/create", article.Create, articleModule.CreateOptions...)
	router.POST(noAuthGroup, "/article/update", article.Update, articleModule.UpdateOptions...)
	router.POST(noAuthGroup, "/article/delete", article.Delete, articleModule.DeleteOptions...)

	//假如这个不用验证
	router.GET(noAuthGroup, "/article/pageList", article.PageList)

	return routeRegistry
}
