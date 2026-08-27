package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysMenuRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	menu := &handler.SysMenu{}
	router.POST(adminApiGroup, "/menu/getMenu", menu.GetMenu)
	router.POST(adminApiGroup, "/menu/getMenuList", menu.GetMenuInfoList)
	router.POST(adminApiGroup, "/menu/getBaseMenuTree", menu.GetBaseMenuTree)
	router.POST(adminApiGroup, "/menu/addBaseMenu", menu.AddBaseMenu)
	router.POST(adminApiGroup, "/menu/deleteBaseMenu", menu.DeleteBaseMenu)
	router.POST(adminApiGroup, "/menu/updateBaseMenu", menu.UpdateBaseMenu)
	router.POST(adminApiGroup, "/menu/getBaseMenuById", menu.GetBaseMenuById)
	router.POST(adminApiGroup, "/menu/getMenuAuthority", menu.GetMenuAuthority)
	router.POST(adminApiGroup, "/menu/addMenuAuthority", menu.AddMenuAuthority)
	router.GET(adminApiGroup, "/menu/getMenuRoles", menu.GetMenuRoles)
	router.POST(adminApiGroup, "/menu/setMenuRoles", menu.SetMenuRoles)

	return adminApiGroup
}
