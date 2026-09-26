package model

// SysSMSConfig 短信开发配置。敏感密钥仅写入，不通过 JSON 返回。
type SysSMSConfig struct {
	Base
	Name            string `gorm:"size:64;not null;comment:配置名称" json:"name"`
	Provider        int    `gorm:"not null;index;comment:短信服务商" json:"provider"`
	AccessKeyID     string `gorm:"size:128;not null;comment:访问密钥ID" json:"access_key_id"`
	AccessKeySecret string `gorm:"size:255;not null;comment:访问密钥Secret" json:"-"`
	Endpoint        string `gorm:"size:255;default:'';comment:服务地址" json:"endpoint"`
	IsDefault       int    `gorm:"default:0;comment:是否默认渠道，0否1是" json:"is_default"`
	Status          int    `gorm:"default:1;index;comment:状态" json:"status"`
	Remark          string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysSMSConfig) TableName() string { return "sys_sms_config" }
