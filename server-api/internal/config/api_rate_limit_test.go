package config

import "testing"

func TestAPIRateLimitDefaults(t *testing.T) {
	conf := &ServiceConfig{}
	conf.applyAPIRateLimitDefaults()

	if !conf.APIRateLimit.Enable || !conf.APIRateLimit.UID.Enable {
		t.Fatal("限流默认规则应该启用")
	}
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
