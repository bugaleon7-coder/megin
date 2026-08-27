package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysOperationRecord struct {
	base.Repository[model.SysOperationRecord]
}

func NewSysOperationRecord(ctx *api.Context) *SysOperationRecord {
	r := &SysOperationRecord{}
	r.Initialize(ctx)
	return r
}

func (r *SysOperationRecord) DeleteByIds(ids []int) error {
	return r.DB().Delete(&model.SysOperationRecord{}, ids).Error
}