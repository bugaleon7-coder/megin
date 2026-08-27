package runtime

import (
	rateLimitModel "megin/internal/system/model"
	"testing"
	"time"
)

func TestRouteRuleOverridesGlobalRule(t *testing.T) {
	rules := []rateLimitModel.RateLimitRule{
		{ID: 1, ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionIP,
			RateCount: 1, IntervalSeconds: 3600, Burst: 1},
		{ID: 2, ScopeType: rateLimitModel.ScopeRoute, HTTPMethod: "GET", RoutePath: "/api/health", Dimension: rateLimitModel.DimensionIP,
			RateCount: 1, IntervalSeconds: 3600, Burst: 2},
	}
	compiled, err := buildSnapshot(rules)
	if err != nil {
		t.Fatalf("构建规则快照失败: %v", err)
	}
	manager := NewManager(nil, time.Minute, time.Minute)
	manager.snapshot.Store(compiled)

	if !manager.AllowIP("GET", "/api/health", "192.0.2.1").Allowed {
		t.Fatal("接口规则的第一个请求应该通过")
	}
	if !manager.AllowIP("GET", "/api/health", "192.0.2.1").Allowed {
		t.Fatal("接口规则应该覆盖burst为1的全局规则，并允许第二个请求")
	}
	if manager.AllowIP("GET", "/api/health", "192.0.2.1").Allowed {
		t.Fatal("接口规则的burst耗尽后应该拒绝第三个请求")
	}

	if !manager.AllowIP("GET", "/api/other", "192.0.2.1").Allowed {
		t.Fatal("没有接口规则时应该回退全局规则")
	}
	if manager.AllowIP("GET", "/api/other", "192.0.2.1").Allowed {
		t.Fatal("全局规则的burst耗尽后应该拒绝第二个请求")
	}
}

func TestBuildSnapshotRejectsDuplicateRules(t *testing.T) {
	rules := []rateLimitModel.RateLimitRule{
		{ID: 1, ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionIP,
			RateCount: 1, IntervalSeconds: 1, Burst: 1},
		{ID: 2, ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionIP,
			RateCount: 2, IntervalSeconds: 1, Burst: 2},
	}
	if _, err := buildSnapshot(rules); err == nil {
		t.Fatal("重复的全局同维度规则应该导致快照构建失败")
	}
}
