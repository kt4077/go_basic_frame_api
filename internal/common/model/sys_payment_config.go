package model

// SysPaymentConfig 支付渠道配置。表：sys_payment_config
type SysPaymentConfig struct {
	Base
	Name         string `gorm:"size:64;not null;comment:配置名称" json:"name"`
	Channel      int    `gorm:"not null;index;comment:支付渠道" json:"channel"`
	AppID        string `gorm:"size:128;not null;comment:应用ID" json:"app_id"`
	MerchantID   string `gorm:"size:128;default:'';comment:商户号" json:"merchant_id"`
	PrivateKey   string `gorm:"type:text;comment:商户私钥" json:"-"`
	PublicKey    string `gorm:"type:text;comment:平台公钥" json:"-"`
	APIv3Key     string `gorm:"size:255;default:'';comment:微信支付APIv3密钥" json:"-"`
	CertSerialNo string `gorm:"size:128;default:'';comment:证书序列号" json:"cert_serial_no"`
	NotifyURL    string `gorm:"size:500;not null;comment:支付回调地址" json:"notify_url"`
	Status       int    `gorm:"default:1;index;comment:状态" json:"status"`
	Sort         int    `gorm:"default:0;comment:排序" json:"sort"`
	Remark       string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysPaymentConfig) TableName() string { return "sys_payment_config" }
