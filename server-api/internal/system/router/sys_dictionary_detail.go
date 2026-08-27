package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysDictionaryDetailRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	detail := &handler.SysDictionaryDetail{}
	router.POST(adminApiGroup, "/sysDictionaryDetail/createSysDictionaryDetail", detail.CreateSysDictionaryDetail)
	router.DELETEWithBody(adminApiGroup, "/sysDictionaryDetail/deleteSysDictionaryDetail", detail.DeleteSysDictionaryDetail)
	router.PUT(adminApiGroup, "/sysDictionaryDetail/updateSysDictionaryDetail", detail.UpdateSysDictionaryDetail)
	router.GET(adminApiGroup, "/sysDictionaryDetail/findSysDictionaryDetail", detail.FindSysDictionaryDetail)
	router.GET(adminApiGroup, "/sysDictionaryDetail/getSysDictionaryDetailList", detail.GetSysDictionaryDetailList)
	router.GET(adminApiGroup, "/sysDictionaryDetail/getDictionaryTreeList", detail.GetDictionaryTreeList)
	router.GET(adminApiGroup, "/sysDictionaryDetail/getDictionaryTreeListByType", detail.GetDictionaryTreeListByType)
	router.GET(adminApiGroup, "/sysDictionaryDetail/getDictionaryDetailsByParent", detail.GetDictionaryDetailsByParent)
	router.GET(adminApiGroup, "/sysDictionaryDetail/getDictionaryPath", detail.GetDictionaryPath)

	return adminApiGroup
}
