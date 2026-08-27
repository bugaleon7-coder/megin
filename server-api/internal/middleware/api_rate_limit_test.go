package middleware

import (
	"encoding/json"
	"fmt"
	"megin/internal/config"
	commonDto "megin/internal/module/common/dto"
	rateLimitModel "megin/internal/module/rate_limit/model"
	rateLimitRuntime "megin/internal/module/rate_limit/runtime"
	"megin/pkg/context/api"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestApiIPRateLimit(t *testing.T) {
	restoreRateLimitConfig(t)
	config.GetConfig().APIRateLimit.Enable = true
	initializeMiddlewareTestManager(t, []rateLimitModel.RateLimitRule{
		{ID: 1, Name: "全局IP", ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionIP,
			RateCount: 1, IntervalSeconds: 3600, Burst: 1, Status: rateLimitModel.StatusEnabled},
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	if err := engine.SetTrustedProxies(nil); err != nil {
		t.Fatalf("设置可信代理失败: %v", err)
	}
	engine.Use(ApiIPRateLimit())
	engine.GET("/api/ping", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })

	first := performRateLimitRequest(engine, "/api/ping", "192.0.2.1:10001", "")
	if first.Code != http.StatusNoContent {
		t.Fatalf("第一次请求状态码错误: %d", first.Code)
	}
	second := performRateLimitRequest(engine, "/api/ping", "192.0.2.1:10002", "")
	assertRateLimited(t, second)
	third := performRateLimitRequest(engine, "/api/ping", "192.0.2.2:10001", "")
	if third.Code != http.StatusNoContent {
		t.Fatalf("不同IP的第一次请求状态码错误: %d", third.Code)
	}
}

func TestApiUIDRateLimit(t *testing.T) {
	restoreRateLimitConfig(t)
	config.GetConfig().APIRateLimit.Enable = true
	initializeMiddlewareTestManager(t, []rateLimitModel.RateLimitRule{
		{ID: 2, Name: "全局UID", ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionUID,
			RateCount: 1, IntervalSeconds: 3600, Burst: 1, Status: rateLimitModel.StatusEnabled},
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		userID := 1
		if ctx.GetHeader("X-Test-UID") == "2" {
			userID = 2
		}
		ctx.Set(commonDto.ApiJwtClaims, &commonDto.Claims{UserID: userID})
		ctx.Next()
	})
	engine.Use(ApiUIDRateLimit())
	engine.GET("/api/user/info", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })

	first := performRateLimitRequest(engine, "/api/user/info", "192.0.2.1:10001", "1")
	if first.Code != http.StatusNoContent {
		t.Fatalf("UID第一次请求状态码错误: %d", first.Code)
	}
	second := performRateLimitRequest(engine, "/api/user/info", "192.0.2.2:10001", "1")
	assertRateLimited(t, second)
	third := performRateLimitRequest(engine, "/api/user/info", "192.0.2.1:10001", "2")
	if third.Code != http.StatusNoContent {
		t.Fatalf("不同UID的第一次请求状态码错误: %d", third.Code)
	}
}

func initializeMiddlewareTestManager(t *testing.T, rules []rateLimitModel.RateLimitRule) {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&rateLimitModel.RateLimitRule{}); err != nil {
		t.Fatalf("创建限流规则测试表失败: %v", err)
	}
	if err := db.Create(&rules).Error; err != nil {
		t.Fatalf("写入限流测试规则失败: %v", err)
	}
	if _, err := rateLimitRuntime.InitDefaultManager(db, time.Minute, time.Minute); err != nil {
		t.Fatalf("初始化限流管理器失败: %v", err)
	}
}

func restoreRateLimitConfig(t *testing.T) {
	t.Helper()
	original := config.GetConfig().APIRateLimit
	t.Cleanup(func() { config.GetConfig().APIRateLimit = original })
}

func performRateLimitRequest(engine http.Handler, path, remoteAddr, userID string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.RemoteAddr = remoteAddr
	if userID != "" {
		request.Header.Set("X-Test-UID", userID)
	}
	engine.ServeHTTP(recorder, request)
	return recorder
}

func assertRateLimited(t *testing.T, recorder *httptest.ResponseRecorder) {
	t.Helper()
	if recorder.Code != http.StatusOK {
		t.Fatalf("限流响应HTTP状态码错误: %d", recorder.Code)
	}
	var result api.Result[any]
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("解析限流响应失败: %v", err)
	}
	if result.Code != 429 || result.Success {
		t.Fatalf("限流响应错误: %+v", result)
	}
}
