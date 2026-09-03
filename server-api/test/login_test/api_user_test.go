package article_test

import (
	"encoding/json/v2"
	"megin/internal"
	"megin/internal/config"
	"megin/internal/dto"
	"megin/pkg/bootstrap"
	"megin/test"
	"sync"
	"testing"
)

var initLoginTestOnce sync.Once

const (
	username = "user1"
	password = "123456"
)

func initLoginTestServer() {
	initLoginTestOnce.Do(func() {
		bootstrap.ServerInitWithMode("../../config/config-dev.yaml", config.RunModeMixed, internal.OnServerStart)
	})
}

// TestApiUserRegister 测试 C 端用户注册接口。
func TestApiUserRegister(t *testing.T) {
	initLoginTestServer()

	resp := test.PostWithoutToken("/api/user/register", dto.RegisterReq{
		LoginName: username,
		Password:  password,
	})
	test.Print(resp.Body.String())
}

// TestApiUserLogin 测试 C 端用户登录接口。
func TestApiUserLogin(t *testing.T) {
	initLoginTestServer()

	resp := test.PostWithoutToken("/api/user/login", dto.LoginReq{
		LoginName: username,
		Password:  password,
	})
	test.Print(resp.Body.String())
}

// TestApiUserInfo 测试 C 端用户信息接口。
func TestApiUserInfo(t *testing.T) {
	initLoginTestServer()

	token := GetToken()
	resp := test.GetWithToken("/api/user/info", token, dto.EmptyReq{})
	test.Print(resp.Body.String())
}

// GetToken 登录并返回 token。
func GetToken() string {
	resp := test.PostWithoutToken("/api/user/login", dto.LoginReq{
		LoginName: username,
		Password:  password,
	})
	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(resp.Body.Bytes(), &result)
	return result.Data.Token
}
