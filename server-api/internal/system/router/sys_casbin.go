package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysCasbinRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	casbin := &handler.SysCasbin{}
	router.POST(adminApiGroup, "/casbin/updateCasbin", casbin.UpdateCasbin)
	router.POST(adminApiGroup, "/casbin/getPolicyPathByAuthorityId", casbin.GetPolicyPathByAuthorityId)

	return adminApiGroup
}
