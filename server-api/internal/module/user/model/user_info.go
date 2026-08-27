package model

import "time"

const TableNameUserInfo = "user_info"

// UserInfo 对应 C 端注册用户账号表。
type UserInfo struct {
	ID             uint       `gorm:"column:id;primaryKey" json:"id"`
	LoginName      string     `gorm:"column:login_name;comment:C端注册用户登录账号" json:"loginName"`
	Password       string     `gorm:"column:password;comment:C端注册用户密码摘要" json:"-"`
	CreatedAt      *time.Time `gorm:"column:created_at;comment:创建时间" json:"createdAt"`
	UpdatedAt      *time.Time `gorm:"column:updated_at;comment:最后更新时间" json:"updatedAt"`
	Salt           string     `gorm:"column:salt;comment:密码盐值" json:"-"`
	LastLoginTime  *time.Time `gorm:"column:last_login_time;comment:最后登录时间" json:"lastLoginTime"`
	MobileAuthCode string     `gorm:"column:mobile_authcode;comment:短信验证码" json:"-"`
	Token          string     `gorm:"column:token;comment:最近一次登录token" json:"token"`
	Mobile         string     `gorm:"column:mobile;comment:手机号" json:"mobile"`
}

func (UserInfo) TableName() string { return TableNameUserInfo }
func (m UserInfo) GetID() any      { return m.ID }
