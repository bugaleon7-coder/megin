package config

import (
	"os"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestAPIRateLimitConfigFromYAML(t *testing.T) {
	content, err := os.ReadFile("../../config/config-dev.yaml")
	if err != nil {
		t.Fatalf("读取开发环境配置失败: %v", err)
	}

	var conf ServiceConfig
	if err := yaml.Unmarshal(content, &conf); err != nil {
		t.Fatalf("解析开发环境配置失败: %v", err)
	}

	if !conf.APIRateLimit.Enable {
		t.Fatal("前台 API 限流应该开启")
	}
	if conf.APIRateLimit.IP.Rate != 2 || conf.APIRateLimit.IP.Burst != 2 {
		t.Fatalf("IP 限流配置错误: %+v", conf.APIRateLimit.IP)
	}
	if !conf.APIRateLimit.UID.Enable {
		t.Fatal("UID 限流应该开启")
	}
	if conf.APIRateLimit.UID.Rate != 0.2 || conf.APIRateLimit.UID.Burst != 1 {
		t.Fatalf("UID 限流配置错误: %+v", conf.APIRateLimit.UID)
	}
	if conf.APIRateLimit.IdleExpirationSeconds != 600 || conf.APIRateLimit.CleanupIntervalSeconds != 300 || conf.APIRateLimit.RefreshIntervalSeconds != 60 {
		t.Fatalf("限流清理配置错误: %+v", conf.APIRateLimit)
	}
}

func TestAPIRateLimitDefaults(t *testing.T) {
	conf := &ServiceConfig{}
	conf.applyAPIRateLimitDefaults()

	if conf.APIRateLimit.IP.Rate != 20 || conf.APIRateLimit.IP.Burst != 40 {
		t.Fatalf("IP 限流默认值错误: %+v", conf.APIRateLimit.IP)
	}
	if conf.APIRateLimit.UID.Rate != 10 || conf.APIRateLimit.UID.Burst != 20 {
		t.Fatalf("UID 限流默认值错误: %+v", conf.APIRateLimit.UID)
	}
	if conf.APIRateLimit.IdleExpirationSeconds != 600 || conf.APIRateLimit.CleanupIntervalSeconds != 300 || conf.APIRateLimit.RefreshIntervalSeconds != 60 {
		t.Fatalf("限流清理默认值错误: %+v", conf.APIRateLimit)
	}
}
