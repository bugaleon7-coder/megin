package repository

import (
	"megin/internal/config"
	rateLimitDto "megin/internal/system/dto"
	rateLimitModel "megin/internal/system/model"
	"megin/pkg/context/api"

	"gorm.io/gorm"
)

// RateLimitRule 负责限流规则的数据访问。
type RateLimitRule struct {
	db *gorm.DB
}

func NewRateLimitRule(ctx *api.Context) *RateLimitRule {
	db := ctx.Tx
	if db == nil {
		db = config.GetMysqlDB()
	}
	return &RateLimitRule{db: db}
}

// NewRateLimitRuleWithDB 创建不依赖 HTTP Context 的规则仓库，供启动加载和定时刷新使用。
func NewRateLimitRuleWithDB(db *gorm.DB) *RateLimitRule {
	return &RateLimitRule{db: db}
}

func (r *RateLimitRule) DB() *gorm.DB { return r.db }

func (r *RateLimitRule) ListEnabled() ([]rateLimitModel.RateLimitRule, error) {
	var rules []rateLimitModel.RateLimitRule
	err := r.db.Where("status = ?", rateLimitModel.StatusEnabled).Order("id ASC").Find(&rules).Error
	return rules, err
}

func (r *RateLimitRule) GetByID(id uint) (rateLimitModel.RateLimitRule, error) {
	var rule rateLimitModel.RateLimitRule
	err := r.db.Where("id = ?", id).First(&rule).Error
	return rule, err
}

func (r *RateLimitRule) Exists(scopeType int, method, path string, dimension int, excludeID uint) (bool, error) {
	query := r.db.Model(&rateLimitModel.RateLimitRule{}).
		Where("scope_type = ? AND http_method = ? AND route_path = ? AND dimension = ?", scopeType, method, path, dimension)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *RateLimitRule) Create(rule *rateLimitModel.RateLimitRule) error {
	return r.db.Create(rule).Error
}

func (r *RateLimitRule) Save(rule *rateLimitModel.RateLimitRule) error {
	return r.db.Save(rule).Error
}

func (r *RateLimitRule) Delete(rule *rateLimitModel.RateLimitRule) error {
	return r.db.Delete(rule).Error
}

func (r *RateLimitRule) PageList(req *rateLimitDto.RateLimitRulePageReq) ([]rateLimitModel.RateLimitRule, int64, error) {
	query := r.db.Model(&rateLimitModel.RateLimitRule{})
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.ScopeType > 0 {
		query = query.Where("scope_type = ?", req.ScopeType)
	}
	if req.Dimension > 0 {
		query = query.Where("dimension = ?", req.Dimension)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.HTTPMethod != "" {
		query = query.Where("http_method = ?", req.HTTPMethod)
	}
	if req.RoutePath != "" {
		query = query.Where("route_path LIKE ?", "%"+req.RoutePath+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	pageNo, pageSize := req.PageNo, req.PageSize
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	var rules []rateLimitModel.RateLimitRule
	err := query.Order("scope_type ASC, route_path ASC, dimension ASC, id ASC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&rules).Error
	return rules, total, err
}
