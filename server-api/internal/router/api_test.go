package router

import (
	"encoding/json"
	"megin/internal/config"
	contextApi "megin/pkg/context/api"
	contextRouter "megin/pkg/context/router"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestHealthRouteWithoutToken 验证健康检查接口注册在免 Token 路由组中。
// 请求不携带 Token 仍应进入 Handler，并返回固定的 ok 数据。
func TestHealthRouteWithoutToken(t *testing.T) {
	originalRateLimit := config.GetConfig().APIRateLimit
	config.GetConfig().APIRateLimit.Enable = false
	t.Cleanup(func() {
		config.GetConfig().APIRateLimit = originalRateLimit
	})

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registry := contextRouter.NewRouteRegistry(engine)
	InitApiRouter(registry)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("健康检查接口 HTTP 状态码错误: %d", recorder.Code)
	}
	var result contextApi.Result[string]
	if err := json.Unmarshal(recorder.Body.Bytes(), &result); err != nil {
		t.Fatalf("解析健康检查接口响应失败: %v", err)
	}
	if result.Code != contextApi.STATUS_SUCCESS || !result.Success || result.Data != "ok" {
		t.Fatalf("健康检查接口响应错误: %+v", result)
	}
}
