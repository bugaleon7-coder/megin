package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysJwtRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	jwt := &handler.SysJwt{}
	router.POST(adminApiGroup, "/jwt/jsonInBlacklist", jwt.JsonInBlacklist)

	return adminApiGroup
}
