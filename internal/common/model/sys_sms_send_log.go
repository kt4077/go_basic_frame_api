package model

import "time"

// SysSMSSendLog 短信发送记录。表：sys_sms_send_log
type SysSMSSendLog struct {
	ID                uint       `gorm:"primaryKey;comment:主键ID" json:"id"`
	CreatedAt         time.Time  `gorm:"comment:创建时间" json:"created_at"`
	ConfigID          uint       `gorm:"not null;index;comment:短信配置ID" json:"config_id"`
	SignatureID       uint       `gorm:"default:0;comment:签名ID" json:"signature_id"`
	TemplateID        uint       `gorm:"default:0;comment:模板ID" json:"template_id"`
	Mobile            string     `gorm:"size:32;not null;index;comment:接收手机号" json:"mobile"`
	Content           string     `gorm:"size:1000;default:'';comment:发送内容" json:"content"`
	Status            int        `gorm:"not null;index;comment:发送状态" json:"status"`
	ProviderMessageID string     `gorm:"size:128;default:'';comment:平台消息ID" json:"provider_message_id"`
	ErrorMessage      string     `gorm:"size:500;default:'';comment:错误信息" json:"error_message"`
	SentAt            *time.Time `gorm:"comment:发送时间" json:"sent_at"`
}

func (SysSMSSendLog) TableName() string { return "sys_sms_send_log" }
