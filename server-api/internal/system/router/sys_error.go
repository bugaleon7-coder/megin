package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysErrorRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	sysError := &handler.SysError{}
	router.DELETE(adminApiGroup, "/sysError/deleteSysError", sysError.DeleteSysError)
	router.DELETE(adminApiGroup, "/sysError/deleteSysErrorByIds", sysError.DeleteSysErrorByIds)
	router.PUT(adminApiGroup, "/sysError/updateSysError", sysError.UpdateSysError)
	router.GET(adminApiGroup, "/sysError/getSysErrorSolution", sysError.GetSysErrorSolution)

	// 无需操作记录
	router.GET(adminApiGroup, "/sysError/findSysError", sysError.FindSysError)
	router.GET(adminApiGroup, "/sysError/getSysErrorList", sysError.GetSysErrorList)

	return adminApiGroup
}

func SysErrorPublicRouter(noAuthGroup *router.RouteGroup) *router.RouteGroup {
	sysError := &handler.SysError{}
	router.POST(noAuthGroup, "/sysError/createSysError", sysError.CreateSysError)
	return noAuthGroup
}
