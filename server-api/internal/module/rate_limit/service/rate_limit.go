package service

import (
	"errors"
	"fmt"
	"math"
	"megin/internal/config"
	rateLimitDto "megin/internal/module/rate_limit/dto"
	rateLimitModel "megin/internal/module/rate_limit/model"
	rateLimitRepo "megin/internal/module/rate_limit/repository"
	"megin/pkg/context/api"
	"megin/pkg/errs"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// RateLimitRule 提供限流规则的业务校验和持久化能力。
type RateLimitRule struct {
	repo *rateLimitRepo.RateLimitRule
}

func NewRateLimitRule(ctx *api.Context) *RateLimitRule {
	return &RateLimitRule{repo: rateLimitRepo.NewRateLimitRule(ctx)}
}

func (s *RateLimitRule) Create(req *rateLimitDto.CreateRateLimitRuleReq, operatorID uint) (rateLimitDto.RateLimitRule, error) {
	method, path, err := normalizeAndValidate(req.ScopeType, req.HTTPMethod, req.RoutePath, req.Dimension, req.RateCount, req.IntervalSeconds, req.Burst)
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	exists, err := s.repo.Exists(req.ScopeType, method, path, req.Dimension, 0)
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	if exists {
		return rateLimitDto.RateLimitRule{}, errs.NewBusinessError(4000, "相同作用范围、接口和维度的限流规则已存在")
	}
	now := time.Now()
	rule := rateLimitModel.RateLimitRule{
		Name: req.Name, ScopeType: req.ScopeType, HTTPMethod: method, RoutePath: path,
		Dimension: req.Dimension, RateCount: req.RateCount, IntervalSeconds: req.IntervalSeconds,
		Burst: req.Burst, Status: req.Status, Remark: req.Remark, CreatedBy: operatorID,
		UpdatedBy: operatorID, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.Create(&rule); err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	return toDTO(rule), nil
}

func (s *RateLimitRule) Update(req *rateLimitDto.UpdateRateLimitRuleReq, operatorID uint) (rateLimitDto.RateLimitRule, error) {
	rule, err := s.repo.GetByID(req.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return rateLimitDto.RateLimitRule{}, errs.NewBusinessError(404, "限流规则不存在")
	}
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	method, path, err := normalizeAndValidate(req.ScopeType, req.HTTPMethod, req.RoutePath, req.Dimension, req.RateCount, req.IntervalSeconds, req.Burst)
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	exists, err := s.repo.Exists(req.ScopeType, method, path, req.Dimension, req.ID)
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	if exists {
		return rateLimitDto.RateLimitRule{}, errs.NewBusinessError(4000, "相同作用范围、接口和维度的限流规则已存在")
	}
	rule.Name = req.Name
	rule.ScopeType = req.ScopeType
	rule.HTTPMethod = method
	rule.RoutePath = path
	rule.Dimension = req.Dimension
	rule.RateCount = req.RateCount
	rule.IntervalSeconds = req.IntervalSeconds
	rule.Burst = req.Burst
	rule.Status = req.Status
	rule.Remark = req.Remark
	rule.UpdatedBy = operatorID
	rule.UpdatedAt = time.Now()
	if err := s.repo.Save(&rule); err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	return toDTO(rule), nil
}

func (s *RateLimitRule) ChangeStatus(req *rateLimitDto.ChangeRateLimitRuleStatusReq, operatorID uint) error {
	rule, err := s.repo.GetByID(req.ID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NewBusinessError(404, "限流规则不存在")
	}
	if err != nil {
		return err
	}
	rule.Status = req.Status
	rule.UpdatedBy = operatorID
	rule.UpdatedAt = time.Now()
	return s.repo.Save(&rule)
}

func (s *RateLimitRule) Delete(id uint) error {
	rule, err := s.repo.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NewBusinessError(404, "限流规则不存在")
	}
	if err != nil {
		return err
	}
	return s.repo.Delete(&rule)
}

func (s *RateLimitRule) Detail(id uint) (rateLimitDto.RateLimitRule, error) {
	rule, err := s.repo.GetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return rateLimitDto.RateLimitRule{}, errs.NewBusinessError(404, "限流规则不存在")
	}
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	return toDTO(rule), nil
}

func (s *RateLimitRule) PageList(req *rateLimitDto.RateLimitRulePageReq) (rateLimitDto.PageResult[rateLimitDto.RateLimitRule], error) {
	req.HTTPMethod = strings.ToUpper(strings.TrimSpace(req.HTTPMethod))
	rules, total, err := s.repo.PageList(req)
	if err != nil {
		return rateLimitDto.PageResult[rateLimitDto.RateLimitRule]{}, err
	}
	items := make([]rateLimitDto.RateLimitRule, len(rules))
	for i := range rules {
		items[i] = toDTO(rules[i])
	}
	pageNo, pageSize := req.PageNo, req.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	totalPage := total / int64(pageSize)
	if total%int64(pageSize) != 0 {
		totalPage++
	}
	return rateLimitDto.PageResult[rateLimitDto.RateLimitRule]{
		PageNo: pageNo, PageSize: pageSize, TotalSize: total, TotalPage: totalPage, List: items,
	}, nil
}

func normalizeAndValidate(scopeType int, method, path string, dimension, rateCount, intervalSeconds, burst int) (string, string, error) {
	if dimension != rateLimitModel.DimensionIP && dimension != rateLimitModel.DimensionUID {
		return "", "", errs.NewBusinessError(4000, "限流维度不合法")
	}
	if rateCount <= 0 || intervalSeconds <= 0 || burst <= 0 {
		return "", "", errs.NewBusinessError(4000, "令牌数量、补充周期和桶容量必须大于0")
	}
	if scopeType == rateLimitModel.ScopeGlobal {
		return "*", "*", nil
	}
	if scopeType != rateLimitModel.ScopeRoute {
		return "", "", errs.NewBusinessError(4000, "规则作用范围不合法")
	}
	method = strings.ToUpper(strings.TrimSpace(method))
	path = strings.TrimSpace(path)
	validMethod := false
	for _, candidate := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch} {
		if method == candidate {
			validMethod = true
			break
		}
	}
	if !validMethod {
		return "", "", errs.NewBusinessError(4000, "HTTP方法不合法")
	}
	if !strings.HasPrefix(path, "/api/") || strings.ContainsAny(path, "?#") {
		return "", "", errs.NewBusinessError(4000, "接口路径必须是以/api/开头的Gin路由模板")
	}
	return method, path, nil
}

func toDTO(rule rateLimitModel.RateLimitRule) rateLimitDto.RateLimitRule {
	return rateLimitDto.RateLimitRule{
		ID: rule.ID, Name: rule.Name, ScopeType: rule.ScopeType, HTTPMethod: rule.HTTPMethod,
		RoutePath: rule.RoutePath, Dimension: rule.Dimension, RateCount: rule.RateCount,
		IntervalSeconds: rule.IntervalSeconds, Burst: rule.Burst, Status: rule.Status,
		Remark: rule.Remark, CreatedBy: rule.CreatedBy, UpdatedBy: rule.UpdatedBy,
		CreatedAt: rule.CreatedAt, UpdatedAt: rule.UpdatedAt,
	}
}

// EnsureSchemaAndSeed 创建规则表，并在空表时写入两条全局规则和健康检查接口示例规则。
func EnsureSchemaAndSeed(db *gorm.DB, conf config.APIRateLimitConfig) error {
	if db == nil {
		return fmt.Errorf("mysql database is nil")
	}
	if err := db.AutoMigrate(&rateLimitModel.RateLimitRule{}); err != nil {
		return err
	}
	var count int64
	if err := db.Model(&rateLimitModel.RateLimitRule{}).Count(&count).Error; err != nil || count > 0 {
		return err
	}
	now := time.Now()
	ipCount, ipInterval := rateToRatio(conf.IP.Rate)
	uidCount, uidInterval := rateToRatio(conf.UID.Rate)
	rules := []rateLimitModel.RateLimitRule{
		{Name: "全局IP限流", ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionIP,
			RateCount: ipCount, IntervalSeconds: ipInterval, Burst: conf.IP.Burst, Status: boolStatus(conf.Enable), CreatedAt: now, UpdatedAt: now},
		{Name: "全局UID限流", ScopeType: rateLimitModel.ScopeGlobal, HTTPMethod: "*", RoutePath: "*", Dimension: rateLimitModel.DimensionUID,
			RateCount: uidCount, IntervalSeconds: uidInterval, Burst: conf.UID.Burst, Status: boolStatus(conf.Enable && conf.UID.Enable), CreatedAt: now, UpdatedAt: now},
		// 健康检查是无需 Token 的公开接口，因此默认只提供接口级 IP 限流示例。
		// 每 5 秒补充 1 个令牌且桶容量为 1，便于开发时直观看到接口规则覆盖全局 IP 规则。
		{Name: "健康检查接口IP限流示例", ScopeType: rateLimitModel.ScopeRoute, HTTPMethod: http.MethodGet, RoutePath: "/api/health", Dimension: rateLimitModel.DimensionIP,
			RateCount: 1, IntervalSeconds: 5, Burst: 1, Status: boolStatus(conf.Enable), Remark: "默认接口级限流示例：每个IP每5秒允许1个请求", CreatedAt: now, UpdatedAt: now},
	}
	return db.Create(&rules).Error
}

func rateToRatio(value float64) (int, int) {
	if value >= 1 && math.Abs(value-math.Round(value)) < 0.000001 {
		return int(math.Round(value)), 1
	}
	for interval := 1; interval <= 10000; interval++ {
		count := value * float64(interval)
		if count >= 1 && math.Abs(count-math.Round(count)) < 0.000001 {
			return int(math.Round(count)), interval
		}
	}
	return int(math.Max(1, math.Round(value*1000))), 1000
}

func boolStatus(enabled bool) int {
	if enabled {
		return rateLimitModel.StatusEnabled
	}
	return rateLimitModel.StatusDisabled
}
