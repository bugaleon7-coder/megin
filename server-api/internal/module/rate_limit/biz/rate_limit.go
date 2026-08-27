package biz

import (
	"fmt"
	rateLimitDto "megin/internal/module/rate_limit/dto"
	rateLimitRuntime "megin/internal/module/rate_limit/runtime"
	rateLimitService "megin/internal/module/rate_limit/service"
	"megin/pkg/context/api"
)

// RateLimitRule 编排规则持久化和运行时快照刷新。
type RateLimitRule struct {
	service *rateLimitService.RateLimitRule
}

func NewRateLimitRule(ctx *api.Context) *RateLimitRule {
	return &RateLimitRule{service: rateLimitService.NewRateLimitRule(ctx)}
}

func (b *RateLimitRule) Create(req *rateLimitDto.CreateRateLimitRuleReq, operatorID uint) (rateLimitDto.RateLimitRule, error) {
	rule, err := b.service.Create(req, operatorID)
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	return rule, reloadAfterMutation()
}

func (b *RateLimitRule) Update(req *rateLimitDto.UpdateRateLimitRuleReq, operatorID uint) (rateLimitDto.RateLimitRule, error) {
	rule, err := b.service.Update(req, operatorID)
	if err != nil {
		return rateLimitDto.RateLimitRule{}, err
	}
	return rule, reloadAfterMutation()
}

func (b *RateLimitRule) ChangeStatus(req *rateLimitDto.ChangeRateLimitRuleStatusReq, operatorID uint) error {
	if err := b.service.ChangeStatus(req, operatorID); err != nil {
		return err
	}
	return reloadAfterMutation()
}

func (b *RateLimitRule) Delete(id uint) error {
	if err := b.service.Delete(id); err != nil {
		return err
	}
	return reloadAfterMutation()
}

func (b *RateLimitRule) Detail(id uint) (rateLimitDto.RateLimitRule, error) {
	return b.service.Detail(id)
}

func (b *RateLimitRule) PageList(req *rateLimitDto.RateLimitRulePageReq) (rateLimitDto.PageResult[rateLimitDto.RateLimitRule], error) {
	return b.service.PageList(req)
}

func (b *RateLimitRule) Refresh() (rateLimitDto.RefreshRateLimitResponse, error) {
	count, err := rateLimitRuntime.ReloadDefault()
	return rateLimitDto.RefreshRateLimitResponse{RuleCount: count}, err
}

func reloadAfterMutation() error {
	if _, err := rateLimitRuntime.ReloadDefault(); err != nil {
		return fmt.Errorf("限流规则已保存，但运行时刷新失败: %w", err)
	}
	return nil
}
