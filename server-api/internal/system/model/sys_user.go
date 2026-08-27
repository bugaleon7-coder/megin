package model

import (
	"megin/internal/base"
	"time"
)

const TableNameSysUser = "sys_users"

type SysUser struct {
	base.SystemModel
	UUID          string         `gorm:"column:uuid;index;comment:用户UUID" json:"uuid"`
	Username      string         `gorm:"column:username;index;comment:用户登录名" json:"userName"`
	Password      string         `gorm:"column:password;comment:用户登录密码" json:"-"`
	NickName      string         `gorm:"column:nick_name;default:系统用户;comment:用户昵称" json:"nickName"`
	HeaderImg     string         `gorm:"column:header_img;default:https://qmplusimg.henrongyi.top/gva_header.jpg;comment:用户头像" json:"headerImg"`
	AuthorityId   uint           `gorm:"column:authority_id;default:888;comment:用户角色ID" json:"authorityId"`
	Phone         string         `gorm:"column:phone;comment:用户手机号" json:"phone"`
	Email         string         `gorm:"column:email;comment:用户邮箱" json:"email"`
	Enable        int            `gorm:"column:enable;default:1;comment:用户是否被冻结 1正常 2冻结" json:"enable"`
	Authority     SysAuthority   `gorm:"foreignKey:AuthorityId;references:AuthorityId" json:"authority"`
	Authorities   []SysAuthority `gorm:"many2many:sys_user_authority;" json:"authorities"`
	OriginSetting JSONMap        `gorm:"type:text;default:null;column:origin_setting;comment:配置" json:"originSetting"`
	TOTPSecret    string         `gorm:"column:totp_secret;type:text;comment:Google TOTP密钥" json:"-"`
	TOTPEnabled   bool           `gorm:"column:totp_enabled;default:false;comment:是否启用Google TOTP" json:"-"`
	TOTPBoundAt   *time.Time     `gorm:"column:totp_bound_at;comment:Google TOTP绑定时间" json:"-"`
}

func (SysUser) TableName() string { return TableNameSysUser }
func (m SysUser) IsNil() bool     { return m.ID == 0 }
func (m SysUser) GetID() any      { return m.ID }
