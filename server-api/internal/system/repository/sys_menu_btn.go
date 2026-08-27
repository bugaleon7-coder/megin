package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
)

type SysMenuBtn struct {
	base.Repository[model.SysBaseMenuBtn]
}

func NewSysMenuBtn(ctx *api.Context) *SysMenuBtn {
	r := &SysMenuBtn{}
	r.Initialize(ctx)
	return r
}

func (r *SysMenuBtn) GetByMenuID(menuID uint) ([]model.SysBaseMenuBtn, error) {
	var btns []model.SysBaseMenuBtn
	err := r.DB().Where("sys_base_menu_id = ?", menuID).Find(&btns).Error
	return btns, err
}