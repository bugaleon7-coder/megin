package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysParams struct {
	base.Repository[model.SysParams]
}

func NewSysParams(ctx *api.Context) *SysParams {
	r := &SysParams{}
	r.Initialize(ctx)
	return r
}

func (r *SysParams) GetByKey(key string) (*model.SysParams, error) {
	var param model.SysParams
	err := r.DB().Where("`key` = ?", key).First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

func (r *SysParams) DeleteByIds(ids []int) error {
	return r.DB().Delete(&model.SysParams{}, ids).Error
}