package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysLoginLog struct {
	base.Repository[model.SysLoginLog]
}

func NewSysLoginLog(ctx *api.Context) *SysLoginLog {
	r := &SysLoginLog{}
	r.Initialize(ctx)
	return r
}

func (r *SysLoginLog) DeleteByIds(ids []int) error {
	return r.DB().Delete(&model.SysLoginLog{}, ids).Error
}