package model

import "time"

const TableNameRateLimitRule = "api_rate_limit_rules"

const (
	ScopeGlobal = 1
	ScopeRoute  = 2

	DimensionIP  = 1
	DimensionUID = 2

	StatusDisabled = 0
	StatusEnabled  = 1
)

// RateLimitRule 保存一条前台 API 限流规则。
// 全局规则的 HTTPMethod 和 RoutePath 固定为 *；接口规则使用真实 HTTP Method 和 Gin 路由模板。
type RateLimitRule struct {
	ID              uint      `gorm:"column:id;primaryKey;autoIncrement;comment:限流规则ID" json:"id"`
	Name            string    `gorm:"column:name;size:100;not null;comment:规则名称" json:"name"`
	ScopeType       int       `gorm:"column:scope_type;not null;uniqueIndex:uk_scope_route_dimension;comment:作用范围 1全局 2指定接口" json:"scope_type"`
	HTTPMethod      string    `gorm:"column:http_method;size:10;not null;default:*;uniqueIndex:uk_scope_route_dimension;comment:HTTP方法" json:"http_method"`
	RoutePath       string    `gorm:"column:route_path;size:255;not null;default:*;uniqueIndex:uk_scope_route_dimension;comment:Gin路由模板" json:"route_path"`
	Dimension       int       `gorm:"column:dimension;not null;uniqueIndex:uk_scope_route_dimension;comment:限流维度 1 IP 2 UID" json:"dimension"`
	RateCount       int       `gorm:"column:rate_count;not null;comment:每个周期补充的令牌数" json:"rate_count"`
	IntervalSeconds int       `gorm:"column:interval_seconds;not null;comment:令牌补充周期秒数" json:"interval_seconds"`
	Burst           int       `gorm:"column:burst;not null;comment:令牌桶容量和瞬时放行上限" json:"burst"`
	Status          int       `gorm:"column:status;not null;default:1;index;comment:状态 0禁用 1启用" json:"status"`
	Remark          string    `gorm:"column:remark;size:500;not null;default:'';comment:备注" json:"remark"`
	CreatedBy       uint      `gorm:"column:created_by;not null;default:0;comment:创建管理员ID" json:"created_by"`
	UpdatedBy       uint      `gorm:"column:updated_by;not null;default:0;comment:最后修改管理员ID" json:"updated_by"`
	CreatedAt       time.Time `gorm:"column:created_at;not null;comment:创建时间" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at;not null;comment:修改时间" json:"updated_at"`
}

func (RateLimitRule) TableName() string { return TableNameRateLimitRule }
func (m RateLimitRule) GetID() any      { return m.ID }
