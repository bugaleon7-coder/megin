package dto

import (
	commonDto "megin/internal/module/common/dto"
	"time"
)

// RateLimitRule 是后台限流规则响应对象。
type RateLimitRule struct {
	ID              uint      `json:"id"`               // 限流规则ID
	Name            string    `json:"name"`             // 规则名称
	ScopeType       int       `json:"scope_type"`       // 作用范围：1全局，2指定接口
	HTTPMethod      string    `json:"http_method"`      // HTTP方法；全局规则为星号
	RoutePath       string    `json:"route_path"`       // Gin路由模板；全局规则为星号
	Dimension       int       `json:"dimension"`        // 限流维度：1 IP，2 UID
	RateCount       int       `json:"rate_count"`       // 每个周期补充的令牌数
	IntervalSeconds int       `json:"interval_seconds"` // 令牌补充周期秒数
	Burst           int       `json:"burst"`            // 令牌桶容量和瞬时放行上限
	Status          int       `json:"status"`           // 状态：0禁用，1启用
	Remark          string    `json:"remark"`           // 规则备注
	CreatedBy       uint      `json:"created_by"`       // 创建管理员ID
	UpdatedBy       uint      `json:"updated_by"`       // 最后修改管理员ID
	CreatedAt       time.Time `json:"created_at"`       // 创建时间
	UpdatedAt       time.Time `json:"updated_at"`       // 修改时间
}

// CreateRateLimitRuleReq 是新增限流规则请求。
type CreateRateLimitRuleReq struct {
	Name            string `json:"name" binding:"required,min=1,max=100"`               // 规则名称
	ScopeType       int    `json:"scope_type" binding:"required,oneof=1 2"`             // 作用范围：1全局，2指定接口
	HTTPMethod      string `json:"http_method" binding:"omitempty,max=10"`              // HTTP方法；接口规则必填，全局规则忽略
	RoutePath       string `json:"route_path" binding:"omitempty,max=255"`              // Gin路由模板；接口规则必填，全局规则忽略
	Dimension       int    `json:"dimension" binding:"required,oneof=1 2"`              // 限流维度：1 IP，2 UID
	RateCount       int    `json:"rate_count" binding:"required,min=1,max=1000000"`     // 每个周期补充的令牌数
	IntervalSeconds int    `json:"interval_seconds" binding:"required,min=1,max=86400"` // 令牌补充周期秒数
	Burst           int    `json:"burst" binding:"required,min=1,max=1000000"`          // 令牌桶容量和瞬时放行上限
	Status          int    `json:"status" binding:"oneof=0 1"`                          // 状态：0禁用，1启用
	Remark          string `json:"remark" binding:"omitempty,max=500"`                  // 规则备注
}

// UpdateRateLimitRuleReq 是修改限流规则请求。
type UpdateRateLimitRuleReq struct {
	ID uint `json:"id" binding:"required,min=1"` // 限流规则ID
	CreateRateLimitRuleReq
}

// RateLimitRuleIDReq 是按ID查询或删除规则的请求。
type RateLimitRuleIDReq struct {
	ID uint `json:"id" form:"id" binding:"required,min=1"` // 限流规则ID
}

// ChangeRateLimitRuleStatusReq 是修改规则状态请求。
type ChangeRateLimitRuleStatusReq struct {
	ID     uint `json:"id" binding:"required,min=1"` // 限流规则ID
	Status int  `json:"status" binding:"oneof=0 1"`  // 新状态：0禁用，1启用
}

// RateLimitRulePageReq 是后台规则分页查询请求。
type RateLimitRulePageReq struct {
	commonDto.PageQuery
	Name       string `form:"name" json:"name" binding:"omitempty,max=100"`               // 按规则名称模糊查询
	ScopeType  int    `form:"scope_type" json:"scope_type" binding:"omitempty,oneof=1 2"` // 按作用范围筛选
	Dimension  int    `form:"dimension" json:"dimension" binding:"omitempty,oneof=1 2"`   // 按限流维度筛选
	Status     *int   `form:"status" json:"status" binding:"omitempty,oneof=0 1"`         // 按启用状态筛选
	HTTPMethod string `form:"http_method" json:"http_method" binding:"omitempty,max=10"`  // 按HTTP方法筛选
	RoutePath  string `form:"route_path" json:"route_path" binding:"omitempty,max=255"`   // 按路由模板模糊查询
}

// RateLimitRulePageResult 是限流规则管理接口的分页响应。
type RateLimitRulePageResult struct {
	PageNo    int             `json:"page_no"`    // 当前页码
	PageSize  int             `json:"page_size"`  // 每页数量
	TotalSize int64           `json:"total_size"` // 总记录数
	TotalPage int64           `json:"total_page"` // 总页数
	List      []RateLimitRule `json:"list"`       // 规则列表
}

// RefreshRateLimitResponse 是手动刷新运行时规则的返回对象。
type RefreshRateLimitResponse struct {
	RuleCount int `json:"rule_count"` // 刷新后内存中启用的规则数量
}
