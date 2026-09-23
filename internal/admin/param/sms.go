package param

type SMSConfigSaveReq struct {
	ID              uint   `json:"id" comment:"主键ID"`
	Name            string `json:"name" binding:"required" comment:"配置名称"`
	Provider        int    `json:"provider" binding:"required,oneof=1 2" comment:"短信服务商"`
	AccessKeyID     string `json:"access_key_id" binding:"required" comment:"访问密钥ID"`
	AccessKeySecret string `json:"access_key_secret" comment:"访问密钥Secret"`
	Endpoint        string `json:"endpoint" comment:"服务地址"`
	Status          int    `json:"status" binding:"required,oneof=1 2" comment:"状态"`
	Remark          string `json:"remark" comment:"备注"`
}

type SMSSignatureSaveReq struct {
	ID       uint   `json:"id" comment:"主键ID"`
	ConfigID uint   `json:"config_id" binding:"required" comment:"短信配置ID"`
	Name     string `json:"name" binding:"required" comment:"签名名称"`
	SignCode string `json:"sign_code" comment:"平台签名编码"`
	Status   int    `json:"status" binding:"required,oneof=1 2" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}

type SMSTemplateSaveReq struct {
	ID           uint   `json:"id" comment:"主键ID"`
	ConfigID     uint   `json:"config_id" binding:"required" comment:"短信配置ID"`
	Name         string `json:"name" binding:"required" comment:"模板名称"`
	TemplateCode string `json:"template_code" binding:"required" comment:"平台模板编码"`
	Type         int    `json:"type" binding:"required,oneof=1 2 3" comment:"模板类型"`
	Content      string `json:"content" comment:"模板内容"`
	Status       int    `json:"status" binding:"required,oneof=1 2" comment:"状态"`
	Remark       string `json:"remark" comment:"备注"`
}

type SMSLogListReq struct {
	Mobile   string `form:"mobile" comment:"手机号"`
	Status   int    `form:"status" comment:"发送状态"`
	Page     int    `form:"page" comment:"页码"`
	PageSize int    `form:"page_size" comment:"每页数量"`
}
