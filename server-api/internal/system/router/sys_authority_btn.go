package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysAuthorityBtnRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	btn := &handler.SysAuthorityBtn{}
	router.GET(adminApiGroup, "/authorityBtn/getAuthorityBtn", btn.GetAuthorityBtn)
	router.POST(adminApiGroup, "/authorityBtn/setAuthorityBtn", btn.SetAuthorityBtn)

	return adminApiGroup
}
