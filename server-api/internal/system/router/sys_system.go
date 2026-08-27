package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysSystemRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	system := &handler.SysSystem{}
	router.POST(adminApiGroup, "/system/getSystemConfig", system.GetSystemConfig)
	router.POST(adminApiGroup, "/system/setSystemConfig", system.SetSystemConfig)
	router.POST(adminApiGroup, "/system/getServerInfo", system.GetServerInfo)
	router.POST(adminApiGroup, "/system/reloadSystem", system.ReloadSystem)

	return adminApiGroup
}
