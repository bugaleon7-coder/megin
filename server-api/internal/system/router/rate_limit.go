package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

// RateLimitRouter 注册系统级 API 限流规则管理接口。
func RateLimitRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	rateLimit := &handler.RateLimitRule{}
	router.POST(adminApiGroup, "/system/rate-limit/create", rateLimit.Create)
	router.PUT(adminApiGroup, "/system/rate-limit/update", rateLimit.Update)
	router.PUT(adminApiGroup, "/system/rate-limit/changeStatus", rateLimit.ChangeStatus)
	router.DELETE(adminApiGroup, "/system/rate-limit/delete", rateLimit.Delete)
	router.GET(adminApiGroup, "/system/rate-limit/detail", rateLimit.Detail)
	router.GET(adminApiGroup, "/system/rate-limit/pageList", rateLimit.PageList)
	router.POST(adminApiGroup, "/system/rate-limit/refresh", rateLimit.Refresh)

	return adminApiGroup
}
