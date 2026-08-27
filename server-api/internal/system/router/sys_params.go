package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysParamsRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	p := &handler.SysParams{}
	router.POST(adminApiGroup, "/sysParams/createSysParams", p.CreateSysParams)
	router.DELETE(adminApiGroup, "/sysParams/deleteSysParams", p.DeleteSysParams)
	router.PUT(adminApiGroup, "/sysParams/updateSysParams", p.UpdateSysParams)
	router.GET(adminApiGroup, "/sysParams/findSysParams", p.GetSysParams)
	router.GET(adminApiGroup, "/sysParams/getSysParamsList", p.GetSysParamsInfoList)
	router.DELETE(adminApiGroup, "/sysParams/deleteSysParamsByIds", p.DeleteSysParamsByIds)
	router.GET(adminApiGroup, "/sysParams/getSysParam", p.GetSysParamsByKey)

	return adminApiGroup
}
