package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysAuthorityBtn struct {
	base.Repository[model.SysAuthorityBtn]
}

func NewSysAuthorityBtn(ctx *api.Context) *SysAuthorityBtn {
	r := &SysAuthorityBtn{}
	r.Initialize(ctx)
	return r
}

func (r *SysAuthorityBtn) GetByAuthority(menuID, authorityId uint) ([]model.SysAuthorityBtn, error) {
	var btns []model.SysAuthorityBtn
	err := r.DB().Where("sys_menu_id = ? AND authority_id = ?", menuID, authorityId).Preload("SysBaseMenuBtn").Find(&btns).Error
	return btns, err
}

func (r *SysAuthorityBtn) DeleteByAuthority(menuID, authorityId uint) error {
	return r.DB().Where("sys_menu_id = ? AND authority_id = ?", menuID, authorityId).Delete(&model.SysAuthorityBtn{}).Error
}

func (r *SysAuthorityBtn) CanRemoveBtn(menuID, btnID, authorityId uint) (bool, error) {
	var count int64
	err := r.DB().Model(&model.SysAuthorityBtn{}).
		Where("sys_base_menu_btn_id = ? AND sys_menu_id != ? AND authority_id = ?", btnID, menuID, authorityId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (r *SysAuthorityBtn) GetAllByBtnID(btnID uint) ([]model.SysAuthorityBtn, error) {
	var btns []model.SysAuthorityBtn
	err := r.DB().Where("sys_base_menu_btn_id = ?", btnID).Find(&btns).Error
	return btns, err
}