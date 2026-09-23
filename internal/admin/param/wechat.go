package param

type WechatConfigSaveReq struct {
	ID        uint   `json:"id" comment:"主键ID"`
	Name      string `json:"name" binding:"required" comment:"配置名称"`
	Type      int    `json:"type" binding:"required,oneof=1 2 3" comment:"微信应用类型"`
	AppID     string `json:"app_id" binding:"required" comment:"微信AppID"`
	AppSecret string `json:"app_secret" comment:"微信AppSecret"`
	Token     string `json:"token" comment:"消息校验Token"`
	AESKey    string `json:"aes_key" comment:"消息加解密密钥"`
	Status    int    `json:"status" binding:"required,oneof=1 2" comment:"状态"`
	Remark    string `json:"remark" comment:"备注"`
}
