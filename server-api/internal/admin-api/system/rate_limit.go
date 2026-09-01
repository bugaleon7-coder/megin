package system

import (
	commonDto "megin/internal/module/common/dto"
	rateLimitBiz "megin/internal/system/biz"
	rateLimitDto "megin/internal/system/dto"
	"megin/pkg/context/api"
)

// RateLimitRule @Tag API限流规则管理
type RateLimitRule struct{}

// Create @Summary 新增API限流规则
// @Description 新增全局规则或指定接口规则；相同接口和维度只允许存在一条规则。
// @Description 请求字段：scope_type为作用范围，dimension为IP或UID维度，rate_count和interval_seconds共同决定补充速率，burst为桶容量。
// @Description 返回字段：返回创建后的完整限流规则。
func (h *RateLimitRule) Create(ctx *api.Context, req *rateLimitDto.CreateRateLimitRuleReq) (*api.Result[rateLimitDto.RateLimitRule], error) {
	rule, err := rateLimitBiz.NewRateLimitRule(ctx).Create(req, uint(ctx.AdminInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(rule)
}

// Update @Summary 修改API限流规则
// @Description 修改规则并在数据库提交后立即刷新当前服务实例的运行时快照。
// @Description 请求字段：id为规则ID，其余字段含义与新增接口一致。
// @Description 返回字段：返回修改后的完整限流规则。
func (h *RateLimitRule) Update(ctx *api.Context, req *rateLimitDto.UpdateRateLimitRuleReq) (*api.Result[rateLimitDto.RateLimitRule], error) {
	rule, err := rateLimitBiz.NewRateLimitRule(ctx).Update(req, uint(ctx.AdminInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(rule)
}

// ChangeStatus @Summary 启用或禁用API限流规则
// @Description 修改状态后立即刷新当前服务实例；禁用接口规则后会回退使用同维度全局规则。
// @Description 请求字段：id为规则ID，status为0禁用或1启用。
// @Description 返回字段：无业务数据。
func (h *RateLimitRule) ChangeStatus(ctx *api.Context, req *rateLimitDto.ChangeRateLimitRuleStatusReq) (*api.Result[Success], error) {
	if err := rateLimitBiz.NewRateLimitRule(ctx).ChangeStatus(req, uint(ctx.AdminInfo.UserID)); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// Delete @Summary 删除API限流规则
// @Description 按规则ID硬删除，并立即刷新当前服务实例的运行时快照。
// @Description 请求字段：id为限流规则ID。
// @Description 返回字段：无业务数据。
func (h *RateLimitRule) Delete(ctx *api.Context, req *rateLimitDto.RateLimitRuleIDReq) (*api.Result[Success], error) {
	if err := rateLimitBiz.NewRateLimitRule(ctx).Delete(req.ID); err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// Detail @Summary 查询API限流规则详情
// @Description 根据规则ID返回完整配置。
// @Description 请求字段：id为限流规则ID。
// @Description 返回字段：返回规则作用范围、接口、维度、速率、桶容量和状态。
func (h *RateLimitRule) Detail(ctx *api.Context, req *rateLimitDto.RateLimitRuleIDReq) (*api.Result[rateLimitDto.RateLimitRule], error) {
	rule, err := rateLimitBiz.NewRateLimitRule(ctx).Detail(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(rule)
}

// PageList @Summary 分页查询API限流规则
// @Description 支持按名称、作用范围、维度、状态、HTTP方法和路由筛选。
// @Description 请求字段：page_no和page_size为分页字段，其余字段为可选筛选条件。
// @Description 返回字段：返回规则列表、总数、页码、每页数量和总页数。
func (h *RateLimitRule) PageList(ctx *api.Context, req *rateLimitDto.RateLimitRulePageReq) (*api.Result[rateLimitDto.RateLimitRulePageResult], error) {
	result, err := rateLimitBiz.NewRateLimitRule(ctx).PageList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}

// Refresh @Summary 手动刷新API限流规则
// @Description 从MySQL重新加载全部启用规则，并原子替换当前实例的运行时规则快照。
// @Description 请求字段：无。
// @Description 返回字段：rule_count为刷新后内存中的启用规则数量。
func (h *RateLimitRule) Refresh(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[rateLimitDto.RefreshRateLimitResponse], error) {
	result, err := rateLimitBiz.NewRateLimitRule(ctx).Refresh()
	if err != nil {
		return nil, err
	}
	return api.ResultData(result)
}
