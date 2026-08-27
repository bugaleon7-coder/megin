package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"
	"strconv"

	"gorm.io/gorm"
)

type SysBaseMenu struct {
	base.Repository[model.SysBaseMenu]
}

func NewSysBaseMenu(ctx *api.Context) *SysBaseMenu {
	r := &SysBaseMenu{}
	r.Initialize(ctx)
	return r
}

func (r *SysBaseMenu) GetMenuTree() ([]model.SysBaseMenu, error) {
	var menus []model.SysBaseMenu
	err := r.DB().Preload("Parameters").Preload("MenuBtn").Where("parent_id = 0").Order("sort ASC").Find(&menus).Error
	if err != nil {
		return nil, err
	}
	for i := range menus {
		if err = r.loadChildren(&menus[i]); err != nil {
			return nil, err
		}
	}
	return menus, nil
}

func (r *SysBaseMenu) loadChildren(menu *model.SysBaseMenu) error {
	var children []model.SysBaseMenu
	err := r.DB().Preload("Parameters").Preload("MenuBtn").Where("parent_id = ?", menu.ID).Order("sort ASC").Find(&children).Error
	if err != nil {
		return err
	}
	for i := range children {
		if err = r.loadChildren(&children[i]); err != nil {
			return err
		}
	}
	menu.Children = children
	return nil
}

func (r *SysBaseMenu) GetByIdWithDetails(id uint) (model.SysBaseMenu, error) {
	var menu model.SysBaseMenu
	err := r.DB().Preload("Parameters").Preload("MenuBtn").Where("id = ?", id).First(&menu).Error
	return menu, err
}

func (r *SysBaseMenu) GetMenuByAuthority(authorityId uint) ([]model.SysBaseMenu, error) {
	var menus []model.SysBaseMenu
	err := r.DB().
		Joins("JOIN sys_authority_menus am ON am.sys_base_menu_id = sys_base_menus.id").
		Where("am.sys_authority_authority_id = ?", authorityId).
		Order("sys_base_menus.sort ASC, sys_base_menus.id ASC").
		Preload("Parameters").
		Preload("MenuBtn").
		Find(&menus).Error
	return menus, err
}

func (r *SysBaseMenu) DeleteMenuAuthorities(menuId uint) error {
	return r.DB().Where("sys_base_menu_id = ?", menuId).Delete(&model.SysAuthorityMenu{}).Error
}

func (r *SysBaseMenu) GetParameters(menuId uint) ([]model.SysBaseMenuParameter, error) {
	var params []model.SysBaseMenuParameter
	err := r.DB().Where("sys_base_menu_id = ?", menuId).Find(&params).Error
	return params, err
}

func (r *SysBaseMenu) GetMenuBtns(menuId uint) ([]model.SysBaseMenuBtn, error) {
	var btns []model.SysBaseMenuBtn
	err := r.DB().Where("sys_base_menu_id = ?", menuId).Find(&btns).Error
	return btns, err
}

// GetAuthoritiesByMenuId 获取拥有指定菜单的角色ID列表
func (r *SysBaseMenu) GetAuthoritiesByMenuId(menuId uint) ([]uint, error) {
	var records []model.SysAuthorityMenu
	err := r.DB().Where("sys_base_menu_id = ?", menuId).Find(&records).Error
	if err != nil {
		return nil, err
	}
	authorityIds := make([]uint, 0, len(records))
	for _, record := range records {
		// MenuId is stored as string in SysAuthorityMenu
		if id, convErr := strconv.ParseUint(record.MenuId, 10, 64); convErr == nil {
			authorityIds = append(authorityIds, uint(id))
		}
	}
	return authorityIds, nil
}

// GetDefaultRouterAuthorityIds 获取将指定菜单设为首页的角色ID列表
func (r *SysBaseMenu) GetDefaultRouterAuthorityIds(menuId uint) ([]uint, error) {
	var authorities []model.SysAuthority
	err := r.DB().Where("default_router = ?", func() string {
		var menu model.SysBaseMenu
		if err := r.DB().First(&menu, menuId).Error; err != nil {
			return ""
		}
		return menu.Path
	}()).Find(&authorities).Error
	if err != nil {
		return nil, err
	}
	authorityIds := make([]uint, 0, len(authorities))
	for _, a := range authorities {
		authorityIds = append(authorityIds, a.AuthorityId)
	}
	return authorityIds, nil
}

// SetMenuAuthorities 全量覆盖某菜单关联的角色列表
func (r *SysBaseMenu) SetMenuAuthorities(menuId uint, authorityIds []uint) error {
	return r.DB().Transaction(func(tx *gorm.DB) error {
		// Delete existing associations
		if err := tx.Where("sys_base_menu_id = ?", menuId).Delete(&model.SysAuthorityMenu{}).Error; err != nil {
			return err
		}
		// Create new associations
		for _, authorityId := range authorityIds {
			record := model.SysAuthorityMenu{
				MenuId:      strconv.Itoa(int(menuId)),
				AuthorityId: strconv.Itoa(int(authorityId)),
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
