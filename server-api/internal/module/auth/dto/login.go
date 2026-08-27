package dto

// RegisterReq 前台注册请求
type RegisterReq struct {
	LoginName string `json:"loginName" binding:"required,min=4,max=64"` // C端注册用户登录账号
	Password  string `json:"password" binding:"required,min=6,max=72"`  // C端注册用户登录密码
}

// LoginReq 前台登录请求
type LoginReq struct {
	LoginName string `json:"loginName" binding:"required"` // C端注册用户登录账号
	Password  string `json:"password" binding:"required"`  // C端注册用户登录密码
}

// LoginUser 前台登录用户信息
type LoginUser struct {
	UID       uint   `json:"uid"`       // C端注册用户ID
	LoginName string `json:"loginName"` // C端注册用户登录账号
	Mobile    string `json:"mobile"`    // C端注册用户手机号
}

// UserInfoResponse 当前登录用户信息响应
type UserInfoResponse struct {
	User LoginUser `json:"user"` // 当前登录的C端用户基础信息
}

// LoginResponse 前台登录响应
type LoginResponse struct {
	User      LoginUser `json:"user"`      // 登录成功后的用户基础信息
	Token     string    `json:"token"`     // 服务端签发的JWT Token
	ExpiresAt int64     `json:"expiresAt"` // Token过期时间，毫秒时间戳
}
