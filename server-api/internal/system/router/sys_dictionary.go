package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysDictionaryRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	dict := &handler.SysDictionary{}
	router.POST(adminApiGroup, "/sysDictionary/createSysDictionary", dict.CreateSysDictionary)
	router.DELETEWithBody(adminApiGroup, "/sysDictionary/deleteSysDictionary", dict.DeleteSysDictionary)
	router.PUT(adminApiGroup, "/sysDictionary/updateSysDictionary", dict.UpdateSysDictionary)
	router.GET(adminApiGroup, "/sysDictionary/findSysDictionary", dict.FindSysDictionary)
	router.GET(adminApiGroup, "/sysDictionary/getSysDictionaryList", dict.GetSysDictionaryList)
	router.GET(adminApiGroup, "/sysDictionary/exportSysDictionary", dict.ExportSysDictionary)
	router.POST(adminApiGroup, "/sysDictionary/importSysDictionary", dict.ImportSysDictionary)

	return adminApiGroup
}
