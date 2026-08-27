package system

import (
	"megin/internal/config"
	systemDto "megin/internal/system/dto"
	systemService "megin/internal/system/service"
	"megin/pkg/captcha"
	"megin/pkg/context/api"
)

// SysUser @Tag 系统用户管理
type SysUser struct{}

// @Summary 创建用户
// @Description 创建系统用户
func (this *SysUser) Register(ctx *api.Context, req *systemDto.RegisterReq) (*api.Result[any], error) {
	_, err := systemService.NewSysUser(ctx).Register(uint(ctx.AdminInfo.RoleId), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 修改密码
// @Description 用户修改自己的密码
func (this *SysUser) ChangePassword(ctx *api.Context, req *systemDto.ChangePasswordReq) (*api.Result[any], error) {
	userId := ctx.AdminInfo.UserID
	err := systemService.NewSysUser(ctx).ChangePassword(uint(userId), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取用户列表
// @Description 分页查询系统用户列表
func (this *SysUser) GetUserList(ctx *api.Context, req *systemDto.GetUserListReq) (*api.Result[systemDto.PageResult[systemDto.SysUser]], error) {
	result, err := systemService.NewSysUser(ctx).GetUserInfoList(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 设置用户角色
// @Description 设置用户的当前角色
func (this *SysUser) SetUserAuthority(ctx *api.Context, req *systemDto.SetUserAuthReq) (*api.Result[any], error) {
	userId := ctx.AdminInfo.UserID
	err := systemService.NewSysUser(ctx).SetUserAuthority(uint(userId), req.AuthorityId)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 设置用户角色列表
// @Description 设置用户的多个角色
func (this *SysUser) SetUserAuthorities(ctx *api.Context, req *systemDto.SetUserAuthoritiesReq) (*api.Result[any], error) {
	err := systemService.NewSysUser(ctx).SetUserAuthorities(uint(ctx.AdminInfo.RoleId), req.ID, req.AuthorityIds)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 删除用户
// @Description 根据用户ID删除用户
func (this *SysUser) DeleteUser(ctx *api.Context, req *systemDto.GetUserInfoReq) (*api.Result[any], error) {
	err := systemService.NewSysUser(ctx).DeleteUser(req.ID, uint(ctx.AdminInfo.UserID), uint(ctx.AdminInfo.RoleId))
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 修改用户信息
// @Description 管理员修改用户信息
func (this *SysUser) SetUserInfo(ctx *api.Context, req *systemDto.ChangeUserInfoReq) (*api.Result[any], error) {
	err := systemService.NewSysUser(ctx).SetUserInfo(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 修改自身信息
// @Description 用户修改自己的信息
func (this *SysUser) SetSelfInfo(ctx *api.Context, req *systemDto.ChangeUserInfoReq) (*api.Result[any], error) {
	userId := uint(ctx.AdminInfo.UserID)
	req.ID = userId
	err := systemService.NewSysUser(ctx).SetSelfInfo(userId, req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 设置用户配置
// @Description 设置用户个人配置(originSetting)
func (this *SysUser) SetSelfSetting(ctx *api.Context, req *systemDto.SetSelfSettingReq) (*api.Result[any], error) {
	userId := uint(ctx.AdminInfo.UserID)
	err := systemService.NewSysUser(ctx).SetSelfSetting(userId, map[string]any(*req))
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取用户信息
// @Description 根据用户ID获取用户详细信息
func (this *SysUser) GetUserInfo(ctx *api.Context, req *systemDto.GetUserInfoReq) (*api.Result[systemDto.SysUserResponse], error) {
	id := req.ID
	if id == 0 {
		id = uint(ctx.AdminInfo.UserID)
	}
	result, err := systemService.NewSysUser(ctx).GetUserInfo(id)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 查找用户
// @Description 根据ID查找用户
func (this *SysUser) FindUserById(ctx *api.Context, req *systemDto.GetUserInfoReq) (*api.Result[systemDto.SysUser], error) {
	user, err := systemService.NewSysUser(ctx).FindUserById(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*user)
}

// @Summary 重置密码
// @Description 管理员重置用户密码
func (this *SysUser) ResetPassword(ctx *api.Context, req *systemDto.ResetPasswordReq) (*api.Result[any], error) {
	err := systemService.NewSysUser(ctx).ResetPassword(uint(ctx.AdminInfo.RoleId), req.ID, req.Password)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取用户角色列表
// @Description 获取指定用户的角色列表
func (this *SysUser) GetUserAuthorities(ctx *api.Context, req *systemDto.GetUserInfoReq) (*api.Result[[]systemDto.SysAuthority], error) {
	authorities, err := systemService.NewSysUser(ctx).GetUserAuthorities(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(authorities)
}

// @Summary 用户登录
// @Description 用户登录，返回JWT Token
func (this *SysUser) Login(ctx *api.Context, req *systemDto.LoginReq) (*api.Result[systemDto.LoginResponse], error) {
	user, token, expiresAt, err := systemService.NewSysUser(ctx).Login(req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.LoginResponse{
		User:      *user,
		Token:     token,
		ExpiresAt: expiresAt,
	})
}

// @Summary 获取登录配置
// @Description 获取登录页所需的最小公开配置
func (this *SysUser) LoginConfig(ctx *api.Context, req *dtoBaseReq) (*api.Result[systemDto.LoginConfigResponse], error) {
	return api.ResultData(systemDto.LoginConfigResponse{
		CaptchaEnabled: config.GetConfig().Admin.Captcha.Enable,
		TOTPEnabled:    config.GetConfig().Admin.TOTP.Enable,
		TOTPIssuer:     config.GetConfig().Admin.TOTP.Issuer,
	})
}

// @Summary 获取Google TOTP状态
// @Description 获取当前登录管理员的Google TOTP绑定状态
func (this *SysUser) GetTOTPStatus(ctx *api.Context, req *dtoBaseReq) (*api.Result[systemDto.TotpStatusResponse], error) {
	result, err := systemService.NewSysUser(ctx).GetTOTPStatus(uint(ctx.AdminInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 初始化Google TOTP
// @Description 为当前登录管理员生成Google TOTP绑定信息
func (this *SysUser) InitTOTP(ctx *api.Context, req *dtoBaseReq) (*api.Result[systemDto.TotpSetupResponse], error) {
	result, err := systemService.NewSysUser(ctx).InitTOTP(uint(ctx.AdminInfo.UserID))
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}

// @Summary 启用Google TOTP
// @Description 校验动态码并启用当前管理员的Google TOTP
func (this *SysUser) EnableTOTP(ctx *api.Context, req *systemDto.TotpCodeReq) (*api.Result[any], error) {
	err := systemService.NewSysUser(ctx).EnableTOTP(uint(ctx.AdminInfo.UserID), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 关闭Google TOTP
// @Description 校验动态码并关闭当前管理员的Google TOTP
func (this *SysUser) DisableTOTP(ctx *api.Context, req *systemDto.TotpCodeReq) (*api.Result[any], error) {
	err := systemService.NewSysUser(ctx).DisableTOTP(uint(ctx.AdminInfo.UserID), req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

// @Summary 获取验证码
// @Description 获取登录验证码
func (this *SysUser) Captcha(ctx *api.Context, req *dtoBaseReq) (*api.Result[systemDto.CaptchaResponse], error) {
	// 验证码关闭时直接告诉前端无需展示输入框，同时避免生成无用图片和内存验证码记录。
	if !config.GetConfig().Admin.Captcha.Enable {
		return api.ResultData(systemDto.CaptchaResponse{OpenCaptcha: false})
	}

	result, err := captcha.Generate()
	if err != nil {
		return nil, err
	}
	return api.ResultData(systemDto.CaptchaResponse{
		CaptchaId:     result.CaptchaId,
		PicPath:       result.PicPath,
		CaptchaLength: result.CaptchaLength,
		OpenCaptcha:   result.OpenCaptcha,
	})
}
