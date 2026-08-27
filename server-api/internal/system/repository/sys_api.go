package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"

	"gorm.io/gorm"
)

type SysApi struct {
	base.Repository[model.SysApi]
}

func NewSysApi(ctx *api.Context) *SysApi {
	r := &SysApi{}
	r.Initialize(ctx)
	return r
}

func (r *SysApi) GetApiGroups() ([]model.SysApi, error) {
	var apis []model.SysApi
	err := r.DB().Find(&apis).Error
	return apis, err
}

func (r *SysApi) GetAllApis() ([]model.SysApi, error) {
	var apis []model.SysApi
	err := r.DB().Find(&apis).Error
	return apis, err
}

func (r *SysApi) DeleteByIds(ids []int) error {
	return r.DB().Delete(&model.SysApi{}, ids).Error
}

func (r *SysApi) GetIgnoreApis() ([]model.SysIgnoreApi, error) {
	var apis []model.SysIgnoreApi
	err := r.DB().Find(&apis).Error
	return apis, err
}

func (r *SysApi) UpsertIgnoreApi(path, method string, flag bool) error {
	if flag {
		ignore := model.SysIgnoreApi{Path: path, Method: method, Flag: true}
		return r.DB().Create(&ignore).Error
	}
	return r.DB().Unscoped().Delete(&model.SysIgnoreApi{}, "path = ? AND method = ?", path, method).Error
}

func (r *SysApi) DeleteByPathAndMethod(path, method string) error {
	return r.DB().Where("path = ? AND method = ?", path, method).Delete(&model.SysApi{}).Error
}

// Transaction wraps a function with a new SysApi repo using the transactional DB
func (r *SysApi) Transaction(fn func(txRepo *SysApi) error) error {
	return r.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := &SysApi{}
		txRepo.EnableTx(&base.TX{DB: tx})
		return fn(txRepo)
	})
}

// EnableTx enables the repository to use a specific DB connection (e.g. for transactions)
func (r *SysApi) EnableTx(tx *base.TX) {
	r.Initialize(&api.Context{Tx: tx.DB})
}
