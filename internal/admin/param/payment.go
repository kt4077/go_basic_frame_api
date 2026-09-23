package param

type PaymentConfigSaveReq struct {
	ID           uint   `json:"id" comment:"主键ID"`
	Name         string `json:"name" binding:"required" comment:"配置名称"`
	Channel      int    `json:"channel" binding:"required,oneof=1 2" comment:"支付渠道"`
	AppID        string `json:"app_id" binding:"required" comment:"应用ID"`
	MerchantID   string `json:"merchant_id" comment:"商户号"`
	PrivateKey   string `json:"private_key" comment:"商户私钥"`
	PublicKey    string `json:"public_key" comment:"平台公钥"`
	APIv3Key     string `json:"api_v3_key" comment:"微信支付APIv3密钥"`
	CertSerialNo string `json:"cert_serial_no" comment:"证书序列号"`
	NotifyURL    string `json:"notify_url" binding:"required,url" comment:"支付回调地址"`
	Status       int    `json:"status" binding:"required,oneof=1 2" comment:"状态"`
	Sort         int    `json:"sort" comment:"排序"`
	Remark       string `json:"remark" comment:"备注"`
}
