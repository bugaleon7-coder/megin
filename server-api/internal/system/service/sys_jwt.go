package service

import (
	"megin/internal/base"
	"megin/internal/config"
	"megin/internal/system/model"
	repo "megin/internal/system/repository"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"time"
)

type SysJwt struct {
	base.Service
	Repo *repo.SysJwtBlacklist
}

func NewSysJwt(ctx *api.Context) *SysJwt {
	s := &SysJwt{}
	s.Initialize(ctx)
	s.Repo = repo.NewSysJwtBlacklist(ctx)
	return s
}

func (s *SysJwt) JsonInBlacklist(jwtStr string) error {
	now := time.Now()
	blacklist := model.JwtBlacklist{
		Jwt: jwtStr,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	return s.Repo.Create(&blacklist)
}

func (s *SysJwt) GetRedisJWT(userID int) (string, error) {
	// In this implementation, we check the database blacklist
	// The actual JWT retrieval from Redis would be implemented separately
	return "", nil
}

func (s *SysJwt) IsBlacklisted(jwtStr string) (bool, error) {
	entry, err := s.Repo.GetByJwt(jwtStr)
	if err != nil {
		return false, err
	}
	return entry != nil && entry.ID > 0, nil
}

func IsBlacklistedJWT(jwtStr string) (bool, error) {
	var count int64
	err := config.GetMysqlDB().Model(&model.JwtBlacklist{}).
		Where("jwt = ?", jwtStr).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
