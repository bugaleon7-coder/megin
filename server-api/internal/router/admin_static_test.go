package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAdminStaticHosting(t *testing.T) {
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		t.Fatalf("创建静态资源目录失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("admin-index"), 0o644); err != nil {
		t.Fatalf("写入后台首页失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "app.js"), []byte("admin-asset"), 0o644); err != nil {
		t.Fatalf("写入后台资源失败: %v", err)
	}

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	registerAdminStatic(engine, dir)

	tests := []struct {
		name     string
		path     string
		status   int
		body     string
		location string
	}{
		{name: "根路径跳转", path: "/admin", status: http.StatusMovedPermanently, location: "/admin/"},
		{name: "后台首页", path: "/admin/", status: http.StatusOK, body: "admin-index"},
		{name: "静态资源", path: "/admin/assets/app.js", status: http.StatusOK, body: "admin-asset"},
		{name: "前端路由回退", path: "/admin/dashboard", status: http.StatusOK, body: "admin-index"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			engine.ServeHTTP(recorder, request)
			if recorder.Code != test.status {
				t.Fatalf("状态码错误: got=%d want=%d", recorder.Code, test.status)
			}
			if test.body != "" && recorder.Body.String() != test.body {
				t.Fatalf("响应内容错误: got=%q want=%q", recorder.Body.String(), test.body)
			}
			if test.location != "" && recorder.Header().Get("Location") != test.location {
				t.Fatalf("跳转地址错误: got=%q want=%q", recorder.Header().Get("Location"), test.location)
			}
		})
	}
}
