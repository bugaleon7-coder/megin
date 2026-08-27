package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysOperationRecordRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	record := &handler.SysOperationRecord{}
	router.DELETEWithBody(adminApiGroup, "/sysOperationRecord/deleteSysOperationRecord", record.DeleteSysOperationRecord)
	router.DELETEWithBody(adminApiGroup, "/sysOperationRecord/deleteSysOperationRecordByIds", record.DeleteSysOperationRecords)
	router.GET(adminApiGroup, "/sysOperationRecord/getSysOperationRecordList", record.GetSysOperationRecordInfoList)

	return adminApiGroup
}
