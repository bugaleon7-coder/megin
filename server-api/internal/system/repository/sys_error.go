package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysError struct {
	base.Repository[model.SysError]
}

func NewSysError(ctx *api.Context) *SysError {
	r := &SysError{}
	r.Initialize(ctx)
	return r
}

func (r *SysError) DeleteByIds(ids []string) error {
	return r.DB().Delete(&model.SysError{}, ids).Error
}
