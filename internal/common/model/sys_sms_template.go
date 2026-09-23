package model

// SysSMSTemplate 短信模板。表：sys_sms_template
type SysSMSTemplate struct {
	Base
	ConfigID     uint   `gorm:"not null;index;comment:短信配置ID" json:"config_id"`
	Name         string `gorm:"size:64;not null;comment:模板名称" json:"name"`
	TemplateCode string `gorm:"size:128;not null;comment:平台模板编码" json:"template_code"`
	Type         int    `gorm:"not null;comment:模板类型" json:"type"`
	Content      string `gorm:"size:500;default:'';comment:模板内容" json:"content"`
	Status       int    `gorm:"default:1;comment:状态" json:"status"`
	Remark       string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysSMSTemplate) TableName() string { return "sys_sms_template" }
