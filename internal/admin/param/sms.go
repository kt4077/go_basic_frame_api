package param

type SMSConfigSaveReq struct {
	ID              uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name            string `json:"name" binding:"required" validate:"配置名称" comment:"配置名称"`
	Provider        int    `json:"provider" binding:"required,oneof=1 2 3 4 5" validate:"短信服务商" comment:"短信服务商：1阿里云、2腾讯云、3短信宝、4 SMS.cn、5云片"`
	AccessKeyID     string `json:"access_key_id" binding:"required" validate:"访问密钥ID" comment:"访问密钥ID"`
	AccessKeySecret string `json:"access_key_secret" validate:"访问密钥Secret" comment:"访问密钥Secret"`
	Endpoint        string `json:"endpoint" validate:"服务地址" comment:"服务地址"`
	IsDefault       int    `json:"is_default" binding:"omitempty,oneof=1" validate:"是否默认渠道" comment:"是否默认渠道：0否，1是"`
	Status          int    `json:"status" binding:"required,oneof=1 2" validate:"状态" comment:"状态"`
	Remark          string `json:"remark" validate:"备注" comment:"备注"`
}

// SMSConfigTestReq 短信渠道测试请求。
type SMSConfigTestReq struct {
	ConfigID uint   `json:"config_id" binding:"required" validate:"短信配置ID" comment:"短信配置ID"`
	Mobile   string `json:"mobile" binding:"required,len=11,numeric,startswith=1" validate:"测试手机号" comment:"接收测试验证码的中国大陆手机号"`
}

type SMSSignatureSaveReq struct {
	ID       uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	ConfigID uint   `json:"config_id" binding:"required" validate:"短信配置ID" comment:"短信配置ID"`
	Name     string `json:"name" binding:"required" validate:"签名名称" comment:"签名名称"`
	SignCode string `json:"sign_code" binding:"required" validate:"平台签名编码" comment:"平台签名编码"`
	Status   int    `json:"status" binding:"required,oneof=1 2" validate:"状态" comment:"状态"`
	Remark   string `json:"remark" validate:"备注" comment:"备注"`
}

type SMSTemplateSaveReq struct {
	ID           uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	ConfigID     uint   `json:"config_id" binding:"required" validate:"短信配置ID" comment:"短信配置ID"`
	Name         string `json:"name" binding:"required" validate:"模板名称" comment:"模板名称"`
	TemplateCode string `json:"template_code" validate:"平台模板编码" comment:"平台模板编码；短信宝、云片可留空"`
	Type         int    `json:"type" binding:"required,oneof=1 2 3" validate:"模板类型" comment:"模板类型"`
	Content      string `json:"content" binding:"required" validate:"模板内容" comment:"模板内容"`
	Status       int    `json:"status" binding:"required,oneof=1 2" validate:"状态" comment:"状态"`
	Remark       string `json:"remark" validate:"备注" comment:"备注"`
}

type SMSLogListReq struct {
	Mobile   string `form:"mobile" validate:"手机号" comment:"手机号"`
	Status   int    `form:"status" validate:"发送状态" comment:"发送状态"`
	Page     int    `form:"page" validate:"页码" comment:"页码"`
	PageSize int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}
