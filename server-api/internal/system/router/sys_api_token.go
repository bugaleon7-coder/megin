package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysApiTokenRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	apiToken := &handler.SysApiToken{}
	router.POST(adminApiGroup, "/sysApiToken/createApiToken", apiToken.CreateApiToken)
	router.POST(adminApiGroup, "/sysApiToken/getApiTokenList", apiToken.GetApiTokenList)
	router.POST(adminApiGroup, "/sysApiToken/deleteApiToken", apiToken.DeleteApiToken)

	return adminApiGroup
}
