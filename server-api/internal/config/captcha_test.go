package config

import "testing"

func TestCaptchaConfigByEnvironment(t *testing.T) {
	// 开发环境关闭图形验证码，方便本地联调时直接使用账号密码登录。
	devConfig := InitConfig("../../config/config-dev.yaml", RunModeMixed)
	if devConfig.Captcha.Enable {
		t.Fatal("开发环境应关闭后台登录图形验证码")
	}

	// 生产环境默认开启图形验证码，避免部署时因为遗漏配置而降低登录保护。
	prodConfig := InitConfig("../../config/config-prod.yaml", RunModeMixed)
	if !prodConfig.Captcha.Enable {
		t.Fatal("生产环境应开启后台登录图形验证码")
	}
}
