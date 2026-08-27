package dto

import (
	"time"
)

// SysUserResponse 用户信息响应
type SysUserResponse struct {
	UserInfo *SysUser `json:"userInfo"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User      SysUser `json:"user"`      // 登录成功后的用户信息
	Token     string  `json:"token"`     // 服务端签发的JWT Token
	ExpiresAt int64   `json:"expiresAt"` // Token过期时间，毫秒时间戳
}

// SysUser 用户DTO
type SysUser struct {
	ID            uint           `json:"ID"`
	UUID          string         `json:"uuid"`
	Username      string         `json:"userName"`
	NickName      string         `json:"nickName"`
	HeaderImg     string         `json:"headerImg"`
	AuthorityId   uint           `json:"authorityId"`
	Phone         string         `json:"phone"`
	Email         string         `json:"email"`
	Enable        int            `json:"enable"`
	CreatedAt     *time.Time     `json:"CreatedAt"`
	UpdatedAt     *time.Time     `json:"UpdatedAt"`
	Authority     SysAuthority   `json:"authority"`
	Authorities   []SysAuthority `json:"authorities"`
	OriginSetting map[string]any `json:"originSetting"`
}

// RegisterReq 注册请求
type RegisterReq struct {
	Username     string `json:"userName" binding:"required,min=4,max=20"`
	Password     string `json:"passWord" binding:"required,min=6,max=20"`
	NickName     string `json:"nickName" binding:"required,max=30"`
	HeaderImg    string `json:"headerImg" binding:"omitempty,url"`
	AuthorityId  uint   `json:"authorityId" binding:"omitempty"`
	Enable       int    `json:"enable" binding:"omitempty,oneof=1 2"`
	AuthorityIds []uint `json:"authorityIds" binding:"omitempty"`
	Phone        string `json:"phone" binding:"omitempty"`
	Email        string `json:"email" binding:"omitempty,email"`
}

// LoginReq 登录请求
type LoginReq struct {
	Username  string `json:"username" binding:"required"`           // 登录用户名
	Password  string `json:"password" binding:"required"`           // 登录密码
	Captcha   string `json:"captcha" binding:"omitempty"`           // 图形验证码结果，api端默认可不传
	CaptchaId string `json:"captchaId" binding:"omitempty"`         // 图形验证码ID，api端默认可不传
	OTP       string `json:"otp" binding:"omitempty,len=6,numeric"` // Google TOTP 六位动态码，启用双因子时必填
}

type LoginConfigResponse struct {
	CaptchaEnabled bool   `json:"captchaEnabled"` // 是否启用后台登录图形验证码
	TOTPEnabled    bool   `json:"totpEnabled"`    // 是否启用 Google TOTP 双因子认证
	TOTPIssuer     string `json:"totpIssuer"`     // Google TOTP 签发方名称
}

type TotpStatusResponse struct {
	Enabled       bool       `json:"enabled"`
	BoundAt       *time.Time `json:"boundAt"`
	Issuer        string     `json:"issuer"`
	Account       string     `json:"account"`
	NeedSetup     bool       `json:"needSetup"`
	SystemEnabled bool       `json:"systemEnabled"`
}

type TotpSetupResponse struct {
	Secret     string `json:"secret"`
	OTPAuthURL string `json:"otpAuthUrl"`
	QRContent  string `json:"qrContent"`
}

type TotpCodeReq struct {
	OTP string `json:"otp" binding:"required,len=6,numeric"`
}

// ChangePasswordReq 修改密码请求
type ChangePasswordReq struct {
	Password    string `json:"password" binding:"required,min=6"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

// ResetPasswordReq 重置密码请求
type ResetPasswordReq struct {
	ID       uint   `json:"id" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SetUserAuthReq 设置用户角色请求
type SetUserAuthReq struct {
	AuthorityId uint `json:"authorityId" binding:"required"`
}

// SetUserAuthoritiesReq 设置用户角色列表请求
type SetUserAuthoritiesReq struct {
	ID           uint   `json:"id" binding:"required"`
	AuthorityIds []uint `json:"authorityIds" binding:"required,min=1"`
}

// ChangeUserInfoReq 修改用户信息请求
type ChangeUserInfoReq struct {
	ID        uint   `json:"id" binding:"required"`
	NickName  string `json:"nickName" binding:"omitempty"`
	Phone     string `json:"phone" binding:"omitempty"`
	Email     string `json:"email" binding:"omitempty,email"`
	HeaderImg string `json:"headerImg" binding:"omitempty,url"`
	Enable    int    `json:"enable" binding:"omitempty,oneof=1 2"`
}

// GetUserListReq 用户列表查询请求
type GetUserListReq struct {
	PageQuery
	Username string `json:"userName" form:"userName" binding:"omitempty"`
	NickName string `json:"nickName" form:"nickName" binding:"omitempty"`
	Phone    string `json:"phone" form:"phone" binding:"omitempty"`
	Email    string `json:"email" form:"email" binding:"omitempty"`
	OrderKey string `json:"orderKey" form:"orderKey" binding:"omitempty"`
	Desc     bool   `json:"desc" form:"desc"`
}

// SetSelfSettingReq 用户配置请求。原版接收根层 JSON 对象，不包裹 setting 字段。
type SetSelfSettingReq map[string]any

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaId     string `json:"captchaId"`     // 验证码唯一标识，登录时通过 captchaId 原样提交
	PicPath       string `json:"picPath"`       // Base64 编码的验证码图片，关闭验证码时为空
	CaptchaLength int    `json:"captchaLength"` // 验证码字符长度，关闭验证码时为零
	OpenCaptcha   bool   `json:"openCaptcha"`   // 是否要求显示并校验图形验证码
}
