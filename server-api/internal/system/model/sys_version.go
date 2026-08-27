package model

import "megin/internal/base"

const TableNameSysVersion = "sys_versions"

// SysVersion 版本管理结构体
type SysVersion struct {
	base.SystemModel
	VersionName *string `gorm:"column:version_name;comment:版本名称;size:255;" json:"versionName"`
	VersionCode *string `gorm:"column:version_code;comment:版本号;size:100;" json:"versionCode"`
	Description *string `gorm:"column:description;comment:版本描述;size:500;" json:"description"`
	VersionData *string `gorm:"column:version_data;comment:版本数据JSON;type:text;" json:"versionData"`
}

func (SysVersion) TableName() string { return TableNameSysVersion }
func (m SysVersion) IsNil() bool     { return m.ID == 0 }
func (m SysVersion) GetID() any      { return m.ID }
