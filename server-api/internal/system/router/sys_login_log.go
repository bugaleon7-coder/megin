package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysLoginLogRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	log := &handler.SysLoginLog{}
	router.DELETEWithBody(adminApiGroup, "/sysLoginLog/deleteLoginLog", log.DeleteSysLoginLog)
	router.DELETEWithBody(adminApiGroup, "/sysLoginLog/deleteLoginLogByIds", log.DeleteSysLoginLogs)
	router.GET(adminApiGroup, "/sysLoginLog/getLoginLogList", log.GetSysLoginLogInfoList)
	router.GET(adminApiGroup, "/sysLoginLog/findLoginLog", log.FindSysLoginLog)

	return adminApiGroup
}
