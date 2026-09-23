package model

// SysSMSSignature 短信签名。表：sys_sms_signature
type SysSMSSignature struct {
	Base
	ConfigID uint   `gorm:"not null;index;comment:短信配置ID" json:"config_id"`
	Name     string `gorm:"size:64;not null;comment:签名名称" json:"name"`
	SignCode string `gorm:"size:128;default:'';comment:平台签名编码" json:"sign_code"`
	Status   int    `gorm:"default:1;comment:状态" json:"status"`
	Remark   string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysSMSSignature) TableName() string { return "sys_sms_signature" }
