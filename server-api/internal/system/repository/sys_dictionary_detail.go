package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysDictionaryDetail struct {
	base.Repository[model.SysDictionaryDetail]
}

func NewSysDictionaryDetail(ctx *api.Context) *SysDictionaryDetail {
	r := &SysDictionaryDetail{}
	r.Initialize(ctx)
	return r
}

func (r *SysDictionaryDetail) GetByDictionaryID(dictID int) ([]model.SysDictionaryDetail, error) {
	var details []model.SysDictionaryDetail
	err := r.DB().Where("sys_dictionary_id = ?", dictID).Order("sort ASC").Find(&details).Error
	return details, err
}

func (r *SysDictionaryDetail) GetChildren(parentID uint) ([]model.SysDictionaryDetail, error) {
	var children []model.SysDictionaryDetail
	err := r.DB().Where("parent_id = ?", parentID).Order("sort ASC").Find(&children).Error
	return children, err
}

func (r *SysDictionaryDetail) GetByParentAndDict(dictID int, parentID *uint) ([]model.SysDictionaryDetail, error) {
	query := r.DB().Where("sys_dictionary_id = ?", dictID).Order("sort ASC")
	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	} else {
		query = query.Where("parent_id IS NULL")
	}
	var details []model.SysDictionaryDetail
	err := query.Find(&details).Error
	return details, err
}
