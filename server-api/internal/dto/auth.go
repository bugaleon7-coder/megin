package dto

type RegisterReq struct {
	LoginName string `json:"loginName" binding:"required,min=4,max=64"` // 注册登录名，长度为 4 至 64 个字符。
	Password  string `json:"password" binding:"required,min=6,max=72"`  // 注册密码，长度为 6 至 72 个字符。
}
type LoginReq struct {
	LoginName string `json:"loginName" binding:"required"` // 登录账号。
	Password  string `json:"password" binding:"required"`  // 登录密码。
}
type LoginUser struct {
	UID       uint   `json:"uid"`       // 用户唯一 ID。
	LoginName string `json:"loginName"` // 登录账号。
	Mobile    string `json:"mobile"`    // 绑定的手机号；未绑定时为空字符串。
}
type UserInfoResponse struct {
	User LoginUser `json:"user"` // 当前登录用户信息。
}
type LoginResponse struct {
	User      LoginUser `json:"user"`      // 登录用户信息。
	Token     string    `json:"token"`     // 服务端签发的访问令牌。
	ExpiresAt int64     `json:"expiresAt"` // 令牌过期时间，Unix 毫秒时间戳。
}
