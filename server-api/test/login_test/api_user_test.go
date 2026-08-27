package article_test

import (
	"encoding/json"
	"fmt"
	"megin/internal"
	"megin/internal/config"
	"megin/pkg/bootstrap"
	"megin/test"
	"strings"
	"sync"
	"testing"
	"time"
)

var initLoginTestOnce sync.Once

func initLoginTestServer() {
	initLoginTestOnce.Do(func() {
		bootstrap.ServerInitWithMode("../../config/config-dev.yaml", config.RunModeMixed, internal.OnServerStart)
	})
}

// TestApiUserRegisterAndLogin 测试 C 端用户注册与登录流程。
func TestApiUserRegisterAndLogin(t *testing.T) {
	initLoginTestServer()

	loginName := fmt.Sprintf("api_user_%d", time.Now().UnixNano())
	password := "12345678"

	registerResp := test.PostWithoutToken("/api/user/register", map[string]string{
		"loginName": loginName,
		"password":  password,
	})

	var registerResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			UID       uint   `json:"uid"`
			LoginName string `json:"loginName"`
			Mobile    string `json:"mobile"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerResp.Body.Bytes(), &registerResult); err != nil {
		t.Fatalf("解析注册响应失败: %v", err)
	}
	if !registerResult.Success {
		t.Fatalf("注册失败: %s", registerResult.Message)
	}
	if registerResult.Data.UID == 0 {
		t.Fatal("注册成功但未返回用户ID")
	}
	if registerResult.Data.LoginName != loginName {
		t.Fatalf("注册返回的登录账号不正确: got=%s want=%s", registerResult.Data.LoginName, loginName)
	}

	var stored struct {
		Password string `gorm:"column:password"`
		Salt     string `gorm:"column:salt"`
	}
	if err := config.GetMysqlDB().
		Table("user_info").
		Select("password", "salt").
		Where("login_name = ?", loginName).
		Take(&stored).Error; err != nil {
		t.Fatalf("查询注册用户失败: %v", err)
	}
	if stored.Salt == "" {
		t.Fatal("注册成功但未生成密码盐值")
	}
	if !strings.HasPrefix(stored.Password, "$2") {
		t.Fatalf("注册成功但密码不是bcrypt格式: %s", stored.Password)
	}

	loginResp := test.PostWithoutToken("/api/user/login", map[string]string{
		"loginName": loginName,
		"password":  password,
	})

	var loginResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			User struct {
				UID       uint   `json:"uid"`
				LoginName string `json:"loginName"`
				Mobile    string `json:"mobile"`
			} `json:"user"`
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginResult); err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
	}
	if !loginResult.Success {
		t.Fatalf("登录失败: %s", loginResult.Message)
	}
	if loginResult.Data.Token == "" {
		t.Fatal("登录成功但未返回token")
	}
	if loginResult.Data.User.UID != registerResult.Data.UID {
		t.Fatalf("登录返回用户ID不匹配: got=%d want=%d", loginResult.Data.User.UID, registerResult.Data.UID)
	}
	if loginResult.Data.User.LoginName != loginName {
		t.Fatalf("登录返回的登录账号不正确: got=%s want=%s", loginResult.Data.User.LoginName, loginName)
	}
	if loginResult.Data.ExpiresAt <= 0 {
		t.Fatalf("登录返回的过期时间不合法: %d", loginResult.Data.ExpiresAt)
	}

	infoResp := test.GetWithToken("/api/user/info", loginResult.Data.Token, nil)
	var infoResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			User struct {
				UID       uint   `json:"uid"`
				LoginName string `json:"loginName"`
				Mobile    string `json:"mobile"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(infoResp.Body.Bytes(), &infoResult); err != nil {
		t.Fatalf("解析用户信息响应失败: %v", err)
	}
	if !infoResult.Success {
		t.Fatalf("获取用户信息失败: %s", infoResult.Message)
	}
	if infoResult.Data.User.UID != registerResult.Data.UID {
		t.Fatalf("用户信息返回ID不匹配: got=%d want=%d", infoResult.Data.User.UID, registerResult.Data.UID)
	}
	if infoResult.Data.User.LoginName != loginName {
		t.Fatalf("用户信息返回登录账号不正确: got=%s want=%s", infoResult.Data.User.LoginName, loginName)
	}
}

// TestApiUserInfoAfterLogin 测试 C 端用户登录后通过 token 获取当前用户信息。
func TestApiUserInfoAfterLogin(t *testing.T) {
	initLoginTestServer()

	loginName := fmt.Sprintf("api_user_info_%d", time.Now().UnixNano())
	password := "12345678"

	registerResp := test.PostWithoutToken("/api/user/register", map[string]string{
		"loginName": loginName,
		"password":  password,
	})

	var registerResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			UID       uint   `json:"uid"`
			LoginName string `json:"loginName"`
			Mobile    string `json:"mobile"`
		} `json:"data"`
	}
	if err := json.Unmarshal(registerResp.Body.Bytes(), &registerResult); err != nil {
		t.Fatalf("解析注册响应失败: %v", err)
	}
	if !registerResult.Success {
		t.Fatalf("注册失败: %s", registerResult.Message)
	}

	loginResp := test.PostWithoutToken("/api/user/login", map[string]string{
		"loginName": loginName,
		"password":  password,
	})

	var loginResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			User struct {
				UID       uint   `json:"uid"`
				LoginName string `json:"loginName"`
				Mobile    string `json:"mobile"`
			} `json:"user"`
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginResult); err != nil {
		t.Fatalf("解析登录响应失败: %v", err)
	}
	if !loginResult.Success {
		t.Fatalf("登录失败: %s", loginResult.Message)
	}
	if loginResult.Data.Token == "" {
		t.Fatal("登录成功但未返回token")
	}

	infoResp := test.GetWithToken("/api/user/info", loginResult.Data.Token, nil)
	var infoResult struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			User struct {
				UID       uint   `json:"uid"`
				LoginName string `json:"loginName"`
				Mobile    string `json:"mobile"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(infoResp.Body.Bytes(), &infoResult); err != nil {
		t.Fatalf("解析用户信息响应失败: %v", err)
	}
	if !infoResult.Success {
		t.Fatalf("获取用户信息失败: %s", infoResult.Message)
	}
	if infoResult.Data.User.UID != registerResult.Data.UID {
		t.Fatalf("用户信息返回ID不匹配: got=%d want=%d", infoResult.Data.User.UID, registerResult.Data.UID)
	}
	if infoResult.Data.User.LoginName != loginName {
		t.Fatalf("用户信息返回登录账号不正确: got=%s want=%s", infoResult.Data.User.LoginName, loginName)
	}
}
