package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysApiRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	api := &handler.SysApi{}
	router.POST(adminApiGroup, "/api/createApi", api.CreateApi)
	router.POST(adminApiGroup, "/api/deleteApi", api.DeleteApi)
	router.POST(adminApiGroup, "/api/updateApi", api.UpdateApi)
	router.POST(adminApiGroup, "/api/getApiById", api.GetApiById)
	router.POST(adminApiGroup, "/api/getApiList", api.GetApiList)
	router.POST(adminApiGroup, "/api/getAllApis", api.GetAllApis)
	router.GET(adminApiGroup, "/api/getApiGroups", api.GetApiGroups)
	router.DELETE(adminApiGroup, "/api/deleteApisByIds", api.DeleteApisByIds)
	router.GET(adminApiGroup, "/api/freshCasbin", api.FreshCasbin)
	router.GET(adminApiGroup, "/api/syncApi", api.SyncApi)
	router.POST(adminApiGroup, "/api/ignoreApi", api.IgnoreApi)
	router.POST(adminApiGroup, "/api/enterSyncApi", api.EnterSyncApi)
	router.GET(adminApiGroup, "/api/getApiRoles", api.GetApiRoles)
	router.POST(adminApiGroup, "/api/setApiRoles", api.SetApiRoles)

	return adminApiGroup
}
