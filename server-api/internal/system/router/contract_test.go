package router

import (
	"reflect"
	"testing"

	sysApi "megin/internal/admin-api/system"
	commonDto "megin/internal/module/common/dto"
	"megin/pkg/context/router"

	"github.com/gin-gonic/gin"
)

// expectedRoute defines a route that must be registered by the system module.
type expectedRoute struct {
	method string
	path   string
}

// contractRoutes lists all routes that the system-api-contract.md requires.
// When adding a new module, append its routes here.
var contractRoutes = []expectedRoute{
	// ── 用户管理 ──
	{method: "POST", path: "/admin-api/user/login"},
	{method: "POST", path: "/admin-api/user/captcha"},
	{method: "GET", path: "/admin-api/user/loginConfig"},
	{method: "POST", path: "/admin-api/user/admin_register"},
	{method: "POST", path: "/admin-api/user/changePassword"},
	{method: "POST", path: "/admin-api/user/getUserList"},
	{method: "POST", path: "/admin-api/user/setUserAuthority"},
	{method: "POST", path: "/admin-api/user/setUserAuthorities"},
	{method: "DELETE", path: "/admin-api/user/deleteUser"},
	{method: "PUT", path: "/admin-api/user/setUserInfo"},
	{method: "PUT", path: "/admin-api/user/setSelfInfo"},
	{method: "PUT", path: "/admin-api/user/setSelfSetting"},
	{method: "GET", path: "/admin-api/user/getUserInfo"},
	{method: "POST", path: "/admin-api/user/resetPassword"},
	{method: "GET", path: "/admin-api/user/getTotpStatus"},
	{method: "POST", path: "/admin-api/user/initTotp"},
	{method: "POST", path: "/admin-api/user/enableTotp"},
	{method: "POST", path: "/admin-api/user/disableTotp"},

	// ── 角色管理 ──
	{method: "POST", path: "/admin-api/authority/getAuthorityList"},
	{method: "POST", path: "/admin-api/authority/createAuthority"},
	{method: "POST", path: "/admin-api/authority/copyAuthority"},
	{method: "POST", path: "/admin-api/authority/deleteAuthority"},
	{method: "PUT", path: "/admin-api/authority/updateAuthority"},
	{method: "POST", path: "/admin-api/authority/setDataAuthority"},
	{method: "GET", path: "/admin-api/authority/getUsersByAuthority"},
	{method: "POST", path: "/admin-api/authority/setRoleUsers"},

	// ── 菜单管理 ──
	{method: "POST", path: "/admin-api/menu/getMenu"},
	{method: "POST", path: "/admin-api/menu/getMenuList"},
	{method: "POST", path: "/admin-api/menu/getBaseMenuTree"},
	{method: "POST", path: "/admin-api/menu/getMenuAuthority"},
	{method: "POST", path: "/admin-api/menu/addMenuAuthority"},
	{method: "POST", path: "/admin-api/menu/addBaseMenu"},
	{method: "POST", path: "/admin-api/menu/deleteBaseMenu"},
	{method: "POST", path: "/admin-api/menu/updateBaseMenu"},
	{method: "POST", path: "/admin-api/menu/getBaseMenuById"},
	{method: "GET", path: "/admin-api/menu/getMenuRoles"},
	{method: "POST", path: "/admin-api/menu/setMenuRoles"},

	// ── API 管理 ──
	{method: "POST", path: "/admin-api/api/getApiList"},
	{method: "POST", path: "/admin-api/api/createApi"},
	{method: "POST", path: "/admin-api/api/getApiById"},
	{method: "POST", path: "/admin-api/api/updateApi"},
	{method: "POST", path: "/admin-api/api/getAllApis"},
	{method: "POST", path: "/admin-api/api/deleteApi"},
	{method: "DELETE", path: "/admin-api/api/deleteApisByIds"},
	{method: "GET", path: "/admin-api/api/getApiGroups"},
	{method: "GET", path: "/admin-api/api/freshCasbin"},
	{method: "GET", path: "/admin-api/api/syncApi"},
	{method: "POST", path: "/admin-api/api/ignoreApi"},
	{method: "POST", path: "/admin-api/api/enterSyncApi"},
	{method: "GET", path: "/admin-api/api/getApiRoles"},
	{method: "POST", path: "/admin-api/api/setApiRoles"},

	// ── Casbin ──
	{method: "POST", path: "/admin-api/casbin/updateCasbin"},
	{method: "POST", path: "/admin-api/casbin/getPolicyPathByAuthorityId"},

	// ── JWT ──
	{method: "POST", path: "/admin-api/jwt/jsonInBlacklist"},

	// ── 参数管理 ──
	{method: "POST", path: "/admin-api/sysParams/createSysParams"},
	{method: "DELETE", path: "/admin-api/sysParams/deleteSysParams"},
	{method: "DELETE", path: "/admin-api/sysParams/deleteSysParamsByIds"},
	{method: "PUT", path: "/admin-api/sysParams/updateSysParams"},
	{method: "GET", path: "/admin-api/sysParams/findSysParams"},
	{method: "GET", path: "/admin-api/sysParams/getSysParamsList"},
	{method: "GET", path: "/admin-api/sysParams/getSysParam"},

	// ── API Token ──
	{method: "POST", path: "/admin-api/sysApiToken/createApiToken"},
	{method: "POST", path: "/admin-api/sysApiToken/getApiTokenList"},
	{method: "POST", path: "/admin-api/sysApiToken/deleteApiToken"},

	// ── 版本管理 ──
	{method: "DELETE", path: "/admin-api/sysVersion/deleteSysVersion"},
	{method: "DELETE", path: "/admin-api/sysVersion/deleteSysVersionByIds"},
	{method: "GET", path: "/admin-api/sysVersion/findSysVersion"},
	{method: "GET", path: "/admin-api/sysVersion/getSysVersionList"},
	{method: "GET", path: "/admin-api/sysVersion/downloadVersionJson"},
	{method: "POST", path: "/admin-api/sysVersion/exportVersion"},
	{method: "POST", path: "/admin-api/sysVersion/importVersion"},
	{method: "GET", path: "/admin-api/sysVersion/getSysVersionPublic"},

	// ── 操作记录 ──
	{method: "DELETE", path: "/admin-api/sysOperationRecord/deleteSysOperationRecord"},
	{method: "DELETE", path: "/admin-api/sysOperationRecord/deleteSysOperationRecordByIds"},
	{method: "GET", path: "/admin-api/sysOperationRecord/getSysOperationRecordList"},

	// ── 登录日志 ──
	{method: "DELETE", path: "/admin-api/sysLoginLog/deleteLoginLog"},
	{method: "DELETE", path: "/admin-api/sysLoginLog/deleteLoginLogByIds"},
	{method: "GET", path: "/admin-api/sysLoginLog/getLoginLogList"},
	{method: "GET", path: "/admin-api/sysLoginLog/findLoginLog"},
}

// collectSystemRoutes builds a fresh engine, registers all system routes,
// and returns the registered RouteInfo slice.
func collectSystemRoutes() []router.RouteInfo {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := router.NewRouteRegistry(engine)
	adminProtectedGroup := registry.Group("admin-api")
	adminProtectedGroup.Use(func(c *gin.Context) {})
	adminPublicGroup := registry.Group("admin-api")

	// Authenticated system routes
	SysUserRouter(adminProtectedGroup)
	SysAuthorityRouter(adminProtectedGroup)
	SysMenuRouter(adminProtectedGroup)
	SysApiRouter(adminProtectedGroup)
	SysCasbinRouter(adminProtectedGroup)
	SysJwtRouter(adminProtectedGroup)
	SysParamsRouter(adminProtectedGroup)
	SysApiTokenRouter(adminProtectedGroup)
	SysVersionRouter(adminProtectedGroup)
	SysOperationRecordRouter(adminProtectedGroup)
	SysLoginLogRouter(adminProtectedGroup)
	SysErrorRouter(adminProtectedGroup)
	SysDictionaryRouter(adminProtectedGroup)
	SysDictionaryDetailRouter(adminProtectedGroup)
	SysSystemRouter(adminProtectedGroup)
	SysAuthorityBtnRouter(adminProtectedGroup)

	// Public routes (login, captcha)
	handler := &sysApi.SysUser{}
	router.POST(adminPublicGroup, "/user/login", handler.Login)
	router.POST(adminPublicGroup, "/user/captcha", handler.Captcha)
	router.GET(adminPublicGroup, "/user/loginConfig", handler.LoginConfig)
	SysErrorPublicRouter(adminPublicGroup)

	return registry.Routes()
}

// buildRouteMap returns a set of "METHOD:path" for quick lookup.
func buildRouteMap(routes []router.RouteInfo) map[string]bool {
	m := make(map[string]bool, len(routes))
	for _, r := range routes {
		key := r.Method + ":" + r.Path
		if m[key] {
			// Duplicate detection will be done in the test
		}
		m[key] = true
	}
	return m
}

func TestContractRoutesExist(t *testing.T) {
	routes := collectSystemRoutes()
	routeMap := buildRouteMap(routes)

	for _, expected := range contractRoutes {
		key := expected.method + ":" + expected.path
		if !routeMap[key] {
			t.Errorf("缺少契约路由: %s %s", expected.method, expected.path)
		}
	}
}

func TestGetApiGroupsAcceptsEmptyQuery(t *testing.T) {
	wantType := reflect.TypeOf((*commonDto.EmptyReq)(nil))
	for _, route := range collectSystemRoutes() {
		if route.Method == "GET" && route.Path == "/admin-api/api/getApiGroups" {
			if gotType := reflect.TypeOf(route.ReqType); gotType != wantType {
				t.Fatalf("getApiGroups 请求类型 = %v, 期望 %v", gotType, wantType)
			}
			return
		}
	}
	t.Fatal("未找到 getApiGroups 路由")
}

func TestNoDuplicateRoutes(t *testing.T) {
	routes := collectSystemRoutes()
	seen := make(map[string]int)
	for _, r := range routes {
		key := r.Method + ":" + r.Path
		if prev, ok := seen[key]; ok {
			t.Errorf("重复路由注册: %s (第 %d 和 %d 条)", key, prev, len(seen)+1)
		}
		seen[key] = len(seen) + 1
	}
}

func TestNoDoubleAdminApiPrefix(t *testing.T) {
	routes := collectSystemRoutes()
	for _, r := range routes {
		if len(r.Path) >= 20 && r.Path[:20] == "/admin-api/admin-api" {
			t.Errorf("路径包含重复前缀: %s %s", r.Method, r.Path)
		}
	}
}

func TestContractRoutesCorrectMethodOnly(t *testing.T) {
	routes := collectSystemRoutes()

	// Group paths by their path portion
	pathMethods := make(map[string][]string)
	for _, r := range routes {
		pathMethods[r.Path] = append(pathMethods[r.Path], r.Method)
	}

	for _, expected := range contractRoutes {
		methods, ok := pathMethods[expected.path]
		if !ok {
			continue // already reported by TestContractRoutesExist
		}
		// Check for extra methods on the same path
		for _, m := range methods {
			if m != expected.method {
				t.Errorf("路径 %s 注册了额外的方法 %s (契约要求: %s)", expected.path, m, expected.method)
			}
		}
	}
}

func TestNoExcludedRoutes(t *testing.T) {
	routes := collectSystemRoutes()
	routeMap := buildRouteMap(routes)

	// Routes that should NOT be registered (deleted from contract)
	excluded := []expectedRoute{
		{method: "POST", path: "/admin-api/user/register"}, // moved to admin_register
		{method: "POST", path: "/admin-api/authority/getAuthorityInfo"},
		{method: "POST", path: "/admin-api/authority/setMenuAuthority"},
		{method: "GET", path: "/admin-api/menu/getMenuTree"},
		{method: "GET", path: "/admin-api/menu/getMenuInfoList"},
		{method: "GET", path: "/admin-api/casbin/clearCasbin"},
		{method: "POST", path: "/admin-api/casbin/updateCasbinApi"},
		{method: "POST", path: "/admin-api/sysParams/getSysParamsInfoList"},
		{method: "GET", path: "/admin-api/sysParams/getSysParamsByKey"},
		{method: "GET", path: "/admin-api/sysParams/getSysParams"},
		{method: "POST", path: "/admin-api/sysOperationRecord/getSysOperationRecordInfoList"},
		{method: "POST", path: "/admin-api/sysOperationRecord/deleteSysOperationRecords"},
		{method: "POST", path: "/admin-api/sysLoginLog/getSysLoginLogInfoList"},
		{method: "POST", path: "/admin-api/sysLoginLog/deleteSysLoginLog"},
		{method: "POST", path: "/admin-api/sysLoginLog/deleteSysLoginLogs"},
	}

	for _, ex := range excluded {
		key := ex.method + ":" + ex.path
		if routeMap[key] {
			t.Errorf("不应注册的旧路由仍存在: %s %s", ex.method, ex.path)
		}
	}
}
