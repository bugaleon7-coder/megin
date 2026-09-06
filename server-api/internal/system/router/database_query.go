package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

// DatabaseQueryRouter 注册数据库只读查询接口。
func DatabaseQueryRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	databaseQuery := &handler.DatabaseQuery{}
	router.GET(adminApiGroup, "/system/database-query/tables", databaseQuery.Tables)
	router.GET(adminApiGroup, "/system/database-query/table-structure", databaseQuery.TableStructure)
	router.POST(adminApiGroup, "/system/database-query/execute", databaseQuery.Execute)
	return adminApiGroup
}
