package param

type PaymentConfigListReq struct {
	Channel int `form:"channel" binding:"omitempty,oneof=1 2" validate:"支付渠道" comment:"支付渠道"`
}

type PaymentConfigSaveReq struct {
	ID           uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name         string `json:"name" binding:"required" validate:"配置名称" comment:"配置名称"`
	Channel      int    `json:"channel" binding:"required,oneof=1 2" validate:"支付渠道" comment:"支付渠道"`
	AppID        string `json:"app_id" binding:"required" validate:"应用ID" comment:"应用ID"`
	MerchantID   string `json:"merchant_id" validate:"商户号" comment:"商户号"`
	PrivateKey   string `json:"private_key" validate:"商户私钥" comment:"商户私钥"`
	PublicKey    string `json:"public_key" validate:"平台公钥" comment:"平台公钥"`
	APIv3Key     string `json:"api_v3_key" validate:"微信支付APIv3密钥" comment:"微信支付APIv3密钥"`
	CertSerialNo string `json:"cert_serial_no" validate:"证书序列号" comment:"证书序列号"`
	NotifyURL    string `json:"notify_url" binding:"required,url" validate:"支付回调地址" comment:"支付回调地址"`
	Status       int    `json:"status" binding:"required,oneof=1 2" validate:"状态" comment:"状态"`
	Sort         int    `json:"sort" validate:"排序" comment:"排序"`
	Remark       string `json:"remark" validate:"备注" comment:"备注"`
}
