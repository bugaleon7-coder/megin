package router

import (
	handler "megin/internal/admin-api/system"
	"megin/pkg/context/router"
)

func SysUserRouter(adminApiGroup *router.RouteGroup) *router.RouteGroup {
	user := &handler.SysUser{}
	router.POST(adminApiGroup, "/user/admin_register", user.Register)
	router.POST(adminApiGroup, "/user/changePassword", user.ChangePassword)
	router.POST(adminApiGroup, "/user/getUserList", user.GetUserList)
	router.GET(adminApiGroup, "/userInfo/getUserInfoList", user.GetUserList)
	router.POST(adminApiGroup, "/user/setUserAuthority", user.SetUserAuthority)
	router.POST(adminApiGroup, "/user/setUserAuthorities", user.SetUserAuthorities)
	router.DELETE(adminApiGroup, "/user/deleteUser", user.DeleteUser)
	router.PUT(adminApiGroup, "/user/setUserInfo", user.SetUserInfo)
	router.PUT(adminApiGroup, "/user/setSelfInfo", user.SetSelfInfo)
	router.PUT(adminApiGroup, "/user/setSelfSetting", user.SetSelfSetting)
	router.GET(adminApiGroup, "/user/getUserInfo", user.GetUserInfo)
	router.POST(adminApiGroup, "/user/resetPassword", user.ResetPassword)
	router.GET(adminApiGroup, "/user/getTotpStatus", user.GetTOTPStatus)
	router.POST(adminApiGroup, "/user/initTotp", user.InitTOTP)
	router.POST(adminApiGroup, "/user/enableTotp", user.EnableTOTP)
	router.POST(adminApiGroup, "/user/disableTotp", user.DisableTOTP)

	return adminApiGroup
}
