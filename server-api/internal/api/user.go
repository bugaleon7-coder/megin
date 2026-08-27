package api

import (
	authBiz "megin/internal/module/auth/biz"
	authDto "megin/internal/module/auth/dto"
	commonDto "megin/internal/module/common/dto"
	"megin/pkg/context/api"
)

// User @Tag 前台用户认证
type User struct{}

// Register @Summary 前台用户注册
// @Description 使用登录账号和登录密码创建一个新的 C 端注册用户。
// @Description 请求字段：loginName 为登录账号，password 为登录密码。
// @Description 返回字段：uid 为新用户ID，loginName 为登录账号，mobile 当前为空字符串。
func (h *User) Register(ctx *api.Context, req *authDto.RegisterReq) (*api.Result[authDto.LoginUser], error) {
	result, err := authBiz.NewAuth(ctx).Register(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// Info @Summary 获取当前登录用户信息
// @Description 该接口必须携带有效的C端用户Token，服务端会根据Token中的uid返回当前登录用户信息。
// @Description 请求字段：无需额外业务字段，仅需在请求头中携带Token。
// @Description 返回字段：user.uid 为用户ID，user.loginName 为登录账号，user.mobile 为手机号。
func (h *User) Info(ctx *api.Context, req *commonDto.EmptyReq) (*api.Result[authDto.UserInfoResponse], error) {
	if ctx.UserInfo == nil {
		return nil, authBiz.NewAuth(ctx).ErrorMessage("用户未登录", 403)
	}
	result, err := authBiz.NewAuth(ctx).GetUserInfo(uint(ctx.UserInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// Login @Summary 前台用户登录
// @Description 使用 C 端注册用户账号登录前台 API。
// @Description 请求字段：loginName 为登录账号，password 为登录密码。
// @Description 返回字段：uid 为用户ID，loginName 为登录账号，mobile 为手机号，token 为访问令牌，expiresAt 为令牌过期时间毫秒时间戳。
func (h *User) Login(ctx *api.Context, req *authDto.LoginReq) (*api.Result[authDto.LoginResponse], error) {
	result, err := authBiz.NewAuth(ctx).Login(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}
