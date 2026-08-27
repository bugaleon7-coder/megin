package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysAuthorityRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	auth := &handler.SysAuthority{}
	router.POST(adminApiGroup, "/authority/createAuthority", auth.CreateAuthority)
	router.POST(adminApiGroup, "/authority/copyAuthority", auth.CopyAuthority)
	router.PUT(adminApiGroup, "/authority/updateAuthority", auth.UpdateAuthority)
	router.POST(adminApiGroup, "/authority/deleteAuthority", auth.DeleteAuthority)
	router.POST(adminApiGroup, "/authority/getAuthorityList", auth.GetAuthorityList)
	router.POST(adminApiGroup, "/authority/setDataAuthority", auth.SetDataAuthority)
	router.GET(adminApiGroup, "/authority/getUsersByAuthority", auth.GetUserIdsByAuthorityId)
	router.POST(adminApiGroup, "/authority/setRoleUsers", auth.SetRoleUsers)

	return adminApiGroup
}
