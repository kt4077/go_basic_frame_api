package model

// SysWechatConfig 微信应用配置。表：sys_wechat_config
type SysWechatConfig struct {
	Base
	Name      string `gorm:"size:64;not null;comment:配置名称" json:"name"`
	Type      int    `gorm:"not null;index;comment:微信应用类型" json:"type"`
	AppID     string `gorm:"size:128;not null;uniqueIndex;comment:微信AppID" json:"app_id"`
	AppSecret string `gorm:"size:255;not null;comment:微信AppSecret" json:"-"`
	Token     string `gorm:"size:255;default:'';comment:消息校验Token" json:"-"`
	AESKey    string `gorm:"size:255;default:'';comment:消息加解密密钥" json:"-"`
	Status    int    `gorm:"default:1;index;comment:状态" json:"status"`
	Remark    string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysWechatConfig) TableName() string { return "sys_wechat_config" }
