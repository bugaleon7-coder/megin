package repository

import (
	"errors"
	"megin/internal/base"
	userModel "megin/internal/module/user/model"
	"megin/pkg/context/api"

	"gorm.io/gorm"
)

// User 负责访问 C 端注册用户账号表。
type User struct {
	base.Repository[userModel.UserInfo]
}

func NewUser(ctx *api.Context) *User {
	r := &User{}
	r.Initialize(ctx)
	return r
}

// GetByLoginName 根据登录账号查询 C 端注册用户。
func (r *User) GetByLoginName(loginName string) (*userModel.UserInfo, error) {
	var user userModel.UserInfo
	err := r.DB().Where("login_name = ?", loginName).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByLoginName 检查指定登录账号是否已存在。
func (r *User) ExistsByLoginName(loginName string) (bool, error) {
	var count int64
	if err := r.DB().Model(&userModel.UserInfo{}).Where("login_name = ?", loginName).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
