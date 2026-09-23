package model

// SysPlatformConfig 平台基础配置。每个平台类型仅保留一条配置记录，
// 后续增加平台级字段时可继续扩展本模型。表：sys_platform_config
type SysPlatformConfig struct {
	Base
	Type            int    `gorm:"not null;uniqueIndex:uk_platform_type;comment:平台类型，1管理端，2用户端" json:"type"`
	Logo            string `gorm:"size:500;default:'';comment:管理端Logo相对路径" json:"logo"`
	SystemName      string `gorm:"size:100;default:'';comment:管理端系统名称" json:"system_name"`
	DefaultNickname string `gorm:"size:100;default:'';comment:用户端默认昵称" json:"default_nickname"`
	DefaultAvatar   string `gorm:"size:500;default:'';comment:用户端默认头像相对路径" json:"default_avatar"`
}

// TableName 返回平台配置表名。
func (SysPlatformConfig) TableName() string { return "sys_platform_config" }
