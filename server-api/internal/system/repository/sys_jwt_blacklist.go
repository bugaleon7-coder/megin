package repository

import (
	"megin/internal/base"
	"megin/internal/system/model"
	"megin/pkg/context/api"

	"gorm.io/gorm"
)

type SysJwtBlacklist struct {
	base.Repository[model.JwtBlacklist]
}

func NewSysJwtBlacklist(ctx *api.Context) *SysJwtBlacklist {
	r := &SysJwtBlacklist{}
	r.Initialize(ctx)
	return r
}

func (r *SysJwtBlacklist) GetByJwt(jwt string) (*model.JwtBlacklist, error) {
	var bl model.JwtBlacklist
	err := r.DB().Where("jwt = ?", jwt).First(&bl).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &bl, nil
}
