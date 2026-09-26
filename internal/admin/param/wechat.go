package param

type WechatConfigListReq struct {
	Type int `form:"type" binding:"omitempty,oneof=1 2 3" validate:"微信应用类型" comment:"微信应用类型"`
}

type WechatConfigSaveReq struct {
	ID        uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name      string `json:"name" binding:"required" validate:"配置名称" comment:"配置名称"`
	Type      int    `json:"type" binding:"required,oneof=1 2 3" validate:"微信应用类型" comment:"微信应用类型"`
	AppID     string `json:"app_id" binding:"required" validate:"微信AppID" comment:"微信AppID"`
	AppSecret string `json:"app_secret" validate:"微信AppSecret" comment:"微信AppSecret"`
	Token     string `json:"token" validate:"消息校验Token" comment:"消息校验Token"`
	AESKey    string `json:"aes_key" validate:"消息加解密密钥" comment:"消息加解密密钥"`
	Status    int    `json:"status" binding:"required,oneof=1 2" validate:"状态" comment:"状态"`
	Remark    string `json:"remark" validate:"备注" comment:"备注"`
}
