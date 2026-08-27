package repository

import (
	"errors"
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"

	"gorm.io/gorm"
)

type SysUser struct {
	base.Repository[model.SysUser]
}

func NewSysUser(ctx *api.Context) *SysUser {
	r := &SysUser{}
	r.Initialize(ctx)
	return r
}

func (r *SysUser) GetByUsername(username string) (*model.SysUser, error) {
	var user model.SysUser
	err := r.DB().Model(&model.SysUser{}).
		Where("username = ?", username).
		Preload("Authority").
		Preload("Authorities").
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *SysUser) GetById(id any) (model.SysUser, error) {
	var user model.SysUser
	err := r.DB().Model(&model.SysUser{}).
		Preload("Authority").
		Preload("Authorities").
		First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, nil
	}
	return user, err
}

func (r *SysUser) GetUserInfoList(db *gorm.DB, pageNo, pageSize int, orderStr string) ([]model.SysUser, int64, error) {
	var users []model.SysUser
	var total int64
	err := db.Model(&model.SysUser{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Limit(pageSize).Offset((pageNo - 1) * pageSize).
		Order(orderStr).
		Preload("Authority").
		Preload("Authorities").
		Find(&users).Error
	return users, total, err
}

func (r *SysUser) DeleteUserAuthorities(userId uint) error {
	return r.DB().Where("sys_user_id = ?", userId).Delete(&model.SysUserAuthority{}).Error
}

func (r *SysUser) SetUserAuthorities(userId uint, authorityIds []uint) error {
	var uas []model.SysUserAuthority
	for _, aid := range authorityIds {
		uas = append(uas, model.SysUserAuthority{
			SysUserId:               userId,
			SysAuthorityAuthorityId: aid,
		})
	}
	return r.DB().Create(&uas).Error
}

func (r *SysUser) GetUserAuthorities(userId uint) ([]model.SysAuthority, error) {
	var authorities []model.SysAuthority
	err := r.DB().Model(&model.SysAuthority{}).
		Select("sys_authorities.*").
		Joins("JOIN sys_user_authority ON sys_user_authority.sys_authority_authority_id = sys_authorities.authority_id").
		Where("sys_user_authority.sys_user_id = ?", userId).
		Find(&authorities).Error
	return authorities, err
}
