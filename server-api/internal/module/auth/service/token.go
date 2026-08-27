package service

import (
	"megin/internal/base"
	"megin/internal/config"
	commonDto "megin/internal/module/common/dto"
	userModel "megin/internal/module/user/model"
	"megin/pkg/context/api"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService 负责生成前台用户登录令牌。
type TokenService struct {
	base.Service
}

func NewTokenService(ctx *api.Context) *TokenService {
	s := &TokenService{}
	s.Initialize(ctx)
	return s
}

// GenerateUserToken 为 C 端注册用户生成 JWT。
func (s *TokenService) GenerateUserToken(user *userModel.UserInfo) (string, int64, error) {
	conf := config.GetConfig()
	expireDuration := time.Duration(conf.Jwt.ExpireSeconds) * time.Second
	if expireDuration <= 0 {
		expireDuration = 24 * time.Hour
	}

	now := time.Now()
	expiresAtTime := now.Add(expireDuration)
	claims := &commonDto.Claims{
		UserID:   int(user.ID),
		Username: user.LoginName,
		Mobile:   user.Mobile,
		RoleId:   0,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAtTime),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(conf.Jwt.Secret))
	if err != nil {
		return "", 0, s.Error(err, "生成前台用户token失败")
	}
	return tokenStr, expiresAtTime.UnixMilli(), nil
}
