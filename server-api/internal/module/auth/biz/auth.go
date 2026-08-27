package biz

import (
	"megin/internal/base"
	authDto "megin/internal/module/auth/dto"
	authService "megin/internal/module/auth/service"
	userService "megin/internal/module/user/service"
	"megin/pkg/context/api"
)

// Auth 负责编排前台用户登录流程。
type Auth struct {
	base.Service
	UserService  *userService.User
	TokenService *authService.TokenService
}

func NewAuth(ctx *api.Context) *Auth {
	s := &Auth{}
	s.Initialize(ctx)
	s.UserService = userService.NewUser(ctx)
	s.TokenService = authService.NewTokenService(ctx)
	return s
}

// Register 执行前台用户注册，并返回注册后的基础信息。
func (s *Auth) Register(req *authDto.RegisterReq) (*authDto.LoginUser, error) {
	user, err := s.UserService.Register(req.LoginName, req.Password)
	if err != nil {
		return nil, err
	}
	return &authDto.LoginUser{
		UID:       user.ID,
		LoginName: user.LoginName,
		Mobile:    user.Mobile,
	}, nil
}

// GetUserInfo 获取当前登录 C 端用户信息。
func (s *Auth) GetUserInfo(userID uint) (*authDto.UserInfoResponse, error) {
	user, err := s.UserService.GetUserInfo(userID)
	if err != nil {
		return nil, err
	}
	return &authDto.UserInfoResponse{
		User: authDto.LoginUser{
			UID:       user.ID,
			LoginName: user.LoginName,
			Mobile:    user.Mobile,
		},
	}, nil
}

// Login 执行前台用户登录，校验账号密码后签发 JWT，并回写最近登录信息。
func (s *Auth) Login(req *authDto.LoginReq) (*authDto.LoginResponse, error) {
	user, err := s.UserService.VerifyLogin(req.LoginName, req.Password)
	if err != nil {
		return nil, err
	}

	token, expiresAt, err := s.TokenService.GenerateUserToken(user)
	if err != nil {
		return nil, err
	}

	if err := s.UserService.AfterLoginSuccess(user.ID, token, expiresAt); err != nil {
		return nil, err
	}

	return &authDto.LoginResponse{
		User: authDto.LoginUser{
			UID:       user.ID,
			LoginName: user.LoginName,
			Mobile:    user.Mobile,
		},
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
