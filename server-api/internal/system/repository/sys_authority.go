package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysAuthority struct {
	base.Repository[model.SysAuthority]
}

func NewSysAuthority(ctx *api.Context) *SysAuthority {
	r := &SysAuthority{}
	r.Initialize(ctx)
	return r
}

func (r *SysAuthority) GetChildren(authorityId uint) ([]model.SysAuthority, error) {
	var list []model.SysAuthority
	err := r.DB().Where("parent_id = ?", authorityId).Find(&list).Error
	return list, err
}

func (r *SysAuthority) GetUserIdsByAuthorityId(authorityId uint) ([]uint, error) {
	var userIds []uint
	err := r.DB().Model(&model.SysUserAuthority{}).
		Where("sys_authority_authority_id = ?", authorityId).
		Pluck("sys_user_id", &userIds).Error
	return userIds, err
}

func (r *SysAuthority) DeleteDataAuthority(authorityId uint) error {
	return r.DB().Where("sys_authority_authority_id = ?", authorityId).Delete(&model.SysAuthorityMenu{}).Error
}

func (r *SysAuthority) SetDataAuthority(authorityId uint, dataAuthorityIds []uint) error {
	// This is handled via GORM many2many, implemented in service
	return nil
}