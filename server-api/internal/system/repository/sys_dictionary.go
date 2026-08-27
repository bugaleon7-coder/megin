package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysDictionary struct {
	base.Repository[model.SysDictionary]
}

func NewSysDictionary(ctx *api.Context) *SysDictionary {
	r := &SysDictionary{}
	r.Initialize(ctx)
	return r
}

func (r *SysDictionary) GetByType(dictType string) (*model.SysDictionary, error) {
	var dict model.SysDictionary
	err := r.DB().Where("type = ?", dictType).Preload("SysDictionaryDetails").First(&dict).Error
	if err != nil {
		return nil, err
	}
	return &dict, nil
}