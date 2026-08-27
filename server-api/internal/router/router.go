package router

import (
	"log"
	"megin/internal/config"
	"megin/internal/middleware"
	"megin/pkg/context/router"
	"megin/pkg/openapi"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// 这里是一个演示中间件
func AuthRequired() gin.HandlerFunc {
	return func(context *gin.Context) {
	}
}

type RouterModules struct {
	MountAPI      bool
	MountAdminAPI bool
}

func InitGinRouter(modules RouterModules) *gin.Engine {
	//初始化数据
	gin.SetMode(gin.DebugMode)
	ginRouter := gin.Default()

	registry := router.NewRouteRegistry(ginRouter)

	registry.Use(middleware.Cors())
	registry.Use(gzip.Gzip(gzip.DefaultCompression))
	registry.Use(middleware.RequestLog())
	registry.Use(middleware.Recover())

	InitGinModules(registry, modules)

	staticRouter(registry)
	conf := config.GetConfig()

	if conf.ApiDoc.Enable {
		// 生成前后台拆分后的 OpenAPI 文档
		genOpenApiDoc(registry, modules)
		staticSwaggerRouter(registry)
		ginRouter.GET("/v3/api-docs/swagger-config", SwaggerConfig)
	}
	return ginRouter
}

func InitGinModules(registry *router.RouteRegistry, modules RouterModules) {
	if modules.MountAPI {
		InitApiRouter(registry)
	}
	if modules.MountAdminAPI {
		adminPublicGroup := NewAdminPublicGroup(registry)
		adminProtectedGroup := NewAdminProtectedGroup(registry)
		AdminApiRouter(adminProtectedGroup)
		InitSystemAdminRouter(adminProtectedGroup)
		InitSystemApiRouter(adminPublicGroup)
	}
}

func genOpenApiDoc(routeRegistry *router.RouteRegistry, modules RouterModules) {
	conf := config.GetConfig()
	if !conf.ApiDoc.Enable {
		return
	}

	if len(conf.ApiDoc.GenScanDir) == 0 {
		conf.ApiDoc.GenScanDir = []string{
			".",
		}
	}

	apiOutputPath := conf.ApiDoc.ApiOutputSwaggerFile
	adminOutputPath := conf.ApiDoc.AdminOutputSwaggerFile
	systemOutputPath := conf.ApiDoc.SystemOutputSwaggerFile
	if apiOutputPath == "" {
		apiOutputPath = conf.ApiDoc.OutputSwaggerFile
	}
	if apiOutputPath == "" {
		apiOutputPath = "static/swagger/api/swagger.json"
	}
	if adminOutputPath == "" {
		adminOutputPath = "static/swagger/admin-api/swagger.json"
	}
	if systemOutputPath == "" {
		systemOutputPath = "static/swagger/system/swagger.json"
	}

	documents := make([]openapi.GenerateOptions, 0, 3)
	if modules.MountAPI {
		documents = append(documents, openapi.GenerateOptions{
			Title:       "API 文档",
			Version:     conf.Version,
			Description: "面向前端与客户端开发的业务 API 文档",
			OutputPath:  apiOutputPath,
			RouteFilter: func(item router.RouteInfo) bool {
				return strings.HasPrefix(item.Path, "/api/")
			},
		})
	}
	if modules.MountAdminAPI {
		documents = append(documents,
			openapi.GenerateOptions{
				Title:       "Admin API 文档",
				Version:     conf.Version,
				Description: "面向后台管理开发的 Admin API 文档",
				OutputPath:  adminOutputPath,
				RouteFilter: func(item router.RouteInfo) bool {
					return strings.HasPrefix(item.Path, "/admin-api/") && !isSystemRoute(item.Path)
				},
			},
			openapi.GenerateOptions{
				Title:       "System API 文档",
				Version:     conf.Version,
				Description: "面向后台系统管理开发的 System API 文档",
				OutputPath:  systemOutputPath,
				RouteFilter: func(item router.RouteInfo) bool {
					return isSystemRoute(item.Path)
				},
			},
		)
	}

	for _, doc := range documents {
		if err := openapi.GenerateOpenAPIDoc(routeRegistry.Routes(), conf.ApiDoc.GenScanDir, doc); err != nil {
			log.Printf("生成 OpenAPI 文档失败 [%s]: %v", doc.OutputPath, err)
			continue
		}
		log.Printf("OpenAPI 文档已生成: %s", doc.OutputPath)
	}
	logSwaggerEntry(conf.Port, modules)
}

// 按当前运行模式输出实际可访问的 Swagger 文档入口，避免启动提示与挂载路由不一致。
func logSwaggerEntry(port string, modules RouterModules) {
	if modules.MountAPI {
		log.Printf("前端 API 文档地址: http://localhost:%s/api-doc/", port)
	}
	if modules.MountAdminAPI {
		log.Printf("后台 Admin API 文档地址: http://localhost:%s/admin-api-doc/", port)
	}
}

func SwaggerConfig(context *gin.Context) {
	docType := detectSwaggerDocType(context)
	switch docType {
	case "admin-api":
		context.JSON(200, buildAdminSwaggerConfig())
		return
	case "admin-system":
		context.JSON(200, buildAdminSystemSwaggerConfig())
		return
	case "mixed":
		context.JSON(200, buildMixedSwaggerConfig())
		return
	default:
		context.JSON(200, buildApiSwaggerConfig())
		return
	}
}

func staticRouter(routeRegistry *router.RouteRegistry) {
	routeRegistry.Engine.Static("api-doc", "./static/knife/")
	routeRegistry.Engine.Static("admin-api-doc", "./static/knife/")
	routeRegistry.Engine.Static("admin-system-doc", "./static/knife/")
}

func staticSwaggerRouter(routeRegistry *router.RouteRegistry) {
	routeRegistry.Engine.Static("swagger", "./static/swagger/")
}

func detectSwaggerDocType(context *gin.Context) string {
	scope := context.Query("scope")
	if scope == "mixed" || scope == "admin-api" || scope == "api" || scope == "admin-system" {
		return scope
	}

	referer := context.Request.Referer()
	switch {
	case strings.Contains(referer, "/admin-api-doc/"):
		return "admin-api"
	case strings.Contains(referer, "/admin-system-doc/"):
		return "admin-system"
	case strings.Contains(referer, "/api-doc/"):
		return "api"
	default:
		switch config.GetConfig().GetRunMode() {
		case config.RunModeAdminAPI:
			return "admin-api"
		case config.RunModeAPI:
			return "api"
		default:
			return "mixed"
		}
	}
}

// 构建 mixed 模式下的 Swagger 配置，直接展示全部文档入口。
func buildMixedSwaggerConfig() gin.H {
	return gin.H{
		"configUrl":         "/v3/api-docs/swagger-config?scope=mixed",
		"oauth2RedirectUrl": "",
		"urls": []gin.H{
			{"name": "api", "url": "/swagger/api/swagger.json"},
			{"name": "admin-api", "url": "/swagger/admin-api/swagger.json"},
			{"name": "admin-system", "url": "/swagger/system/swagger.json"},
		},
		"validatorUrl": "",
	}
}

// 构建前台业务 API 的 Swagger 配置。
func buildApiSwaggerConfig() gin.H {
	return gin.H{
		"configUrl":         "/swagger/api/swagger.json",
		"oauth2RedirectUrl": "",
		"url":               "/swagger/api/swagger.json",
		"validatorUrl":      "",
	}
}

// 构建后台管理 API 的 Swagger 配置，同时携带系统管理文档入口。
func buildAdminSwaggerConfig() gin.H {
	return gin.H{
		"configUrl":         "/v3/api-docs/swagger-config?scope=admin-api",
		"oauth2RedirectUrl": "",
		"urls": []gin.H{
			{"name": "admin-api", "url": "/swagger/admin-api/swagger.json"},
			{"name": "admin-system", "url": "/swagger/system/swagger.json"},
		},
		"validatorUrl": "",
	}
}

// 构建系统管理模块的 Swagger 配置。
func buildAdminSystemSwaggerConfig() gin.H {
	return gin.H{
		"configUrl":         "/swagger/system/swagger.json",
		"oauth2RedirectUrl": "",
		"url":               "/swagger/system/swagger.json",
		"validatorUrl":      "",
	}
}

var systemRoutePrefixes = []string{
	"/admin-api/user/",
	"/admin-api/userInfo/",
	"/admin-api/authority/",
	"/admin-api/authorityBtn/",
	"/admin-api/menu/",
	"/admin-api/api/",
	"/admin-api/casbin/",
	"/admin-api/jwt/",
	"/admin-api/sysParams/",
	"/admin-api/sysApiToken/",
	"/admin-api/sysVersion/",
	"/admin-api/sysOperationRecord/",
	"/admin-api/sysLoginLog/",
	"/admin-api/sysDictionary/",
	"/admin-api/sysDictionaryDetail/",
	"/admin-api/sysError/",
	"/admin-api/system/",
}

func isSystemRoute(path string) bool {
	for _, prefix := range systemRoutePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
