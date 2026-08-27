package article_test

import (
	"encoding/json"
	"fmt"
	"megin/internal/config"
	"megin/pkg/bootstrap"
	"megin/test"
	"testing"

	"megin/internal"
)

// TestAdminLogin 测试管理后台登录
func TestAdminLogin(t *testing.T) {
	// 初始化服务（TestMain已经会在包初始化时执行，但这里独立测试也需要初始化）
	bootstrap.ServerInitWithMode("../../config/config-dev.yaml", config.RunModeMixed, internal.OnServerStart)

	// 登录 admin/123456
	resp := test.PostWithoutToken("/admin-api/user/login", map[string]string{
		"username": "admin",
		"password": "123456",
	})

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			User      any    `json:"user"`
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"data"`
	}

	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	// 打印完整响应
	test.Print(resp.Body.String())

	// 验证登录成功
	if !result.Success {
		t.Fatalf("登录失败: %s", result.Message)
	}

	if result.Data.Token == "" {
		t.Fatal("登录成功但未返回token")
	}

	fmt.Printf("登录成功，token: %s\n", result.Data.Token)
	fmt.Printf("过期时间: %d\n", result.Data.ExpiresAt)
}
