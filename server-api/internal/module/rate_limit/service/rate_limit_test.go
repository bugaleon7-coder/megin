package service

import (
	"fmt"
	"megin/internal/config"
	rateLimitModel "megin/internal/module/rate_limit/model"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestNormalizeAndValidateRouteRule 验证接口规则会统一 HTTP 方法，并拒绝非前台 API 路径。
func TestNormalizeAndValidateRouteRule(t *testing.T) {
	method, path, err := normalizeAndValidate(
		rateLimitModel.ScopeRoute,
		" get ",
		"/api/health",
		rateLimitModel.DimensionIP,
		1,
		5,
		1,
	)
	if err != nil {
		t.Fatalf("合法接口规则校验失败: %v", err)
	}
	if method != "GET" || path != "/api/health" {
		t.Fatalf("接口规则标准化结果错误: method=%s path=%s", method, path)
	}

	if _, _, err := normalizeAndValidate(
		rateLimitModel.ScopeRoute,
		"GET",
		"/admin-api/rate-limit/pageList",
		rateLimitModel.DimensionIP,
		1,
		5,
		1,
	); err == nil {
		t.Fatal("后台接口不应该允许配置为前台 API 限流规则")
	}
}

// TestEnsureSchemaAndSeed 验证首次启动会创建规则表，并写入全局规则和健康检查接口示例规则。
func TestEnsureSchemaAndSeed(t *testing.T) {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("初始化测试数据库失败: %v", err)
	}
	conf := config.APIRateLimitConfig{
		Enable: true,
		IP:     config.RateLimitRule{Rate: 2, Burst: 2},
		UID: config.UIDRateLimitConfig{
			Enable:        true,
			RateLimitRule: config.RateLimitRule{Rate: 0.2, Burst: 1},
		},
	}
	if err := EnsureSchemaAndSeed(db, conf); err != nil {
		t.Fatalf("创建并初始化限流规则表失败: %v", err)
	}

	var rules []rateLimitModel.RateLimitRule
	if err := db.Order("id ASC").Find(&rules).Error; err != nil {
		t.Fatalf("查询初始限流规则失败: %v", err)
	}
	if len(rules) != 3 {
		t.Fatalf("初始规则数量错误: got=%d want=3", len(rules))
	}
	if rules[0].RateCount != 2 || rules[0].IntervalSeconds != 1 || rules[0].Burst != 2 {
		t.Fatalf("全局 IP 初始规则错误: %+v", rules[0])
	}
	if rules[1].RateCount != 1 || rules[1].IntervalSeconds != 5 || rules[1].Burst != 1 {
		t.Fatalf("全局 UID 初始规则错误: %+v", rules[1])
	}
	if rules[2].ScopeType != rateLimitModel.ScopeRoute || rules[2].HTTPMethod != "GET" || rules[2].RoutePath != "/api/health" ||
		rules[2].Dimension != rateLimitModel.DimensionIP || rules[2].RateCount != 1 || rules[2].IntervalSeconds != 5 || rules[2].Burst != 1 {
		t.Fatalf("健康检查接口初始规则错误: %+v", rules[2])
	}

	// 再次执行启动初始化不得重复插入规则。
	if err := EnsureSchemaAndSeed(db, conf); err != nil {
		t.Fatalf("重复初始化限流规则表失败: %v", err)
	}
	var count int64
	if err := db.Model(&rateLimitModel.RateLimitRule{}).Count(&count).Error; err != nil {
		t.Fatalf("统计限流规则失败: %v", err)
	}
	if count != 3 {
		t.Fatalf("重复初始化后规则数量错误: got=%d want=3", count)
	}
}
