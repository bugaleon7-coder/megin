package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

// PprofRouter 注册受后台登录权限保护的性能分析接口。
func PprofRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	pprof := &handler.Pprof{}
	router.GET(adminApiGroup, "/system/pprof/status", pprof.Status)
	router.POST(adminApiGroup, "/system/pprof/toggle", pprof.Toggle)
	router.POST(adminApiGroup, "/system/pprof/cpu/start", pprof.StartCPUProfile)
	router.GET(adminApiGroup, "/system/pprof/cpu/result", pprof.CPUProfileResult)
	router.GET(adminApiGroup, "/system/pprof/records", pprof.Records)
	router.GET(adminApiGroup, "/system/pprof/record", pprof.Record)
	return adminApiGroup
}
