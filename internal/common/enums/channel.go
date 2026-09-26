package enums

// 短信服务商，枚举从 1 开始。
const (
	SMSProviderAliyun  = 1 // 阿里云短信
	SMSProviderTencent = 2 // 腾讯云短信
	SMSProviderSMSBao  = 3 // 短信宝
	SMSProviderSMSCN   = 4 // SMS.cn
	SMSProviderYunpian = 5 // 云片
)

// 短信模板类型，枚举从 1 开始。
const (
	SMSTemplateVerifyCode = 1
	SMSTemplateNotice     = 2
	SMSTemplateMarketing  = 3
)

// 短信发送状态，枚举从 1 开始。
const (
	SMSSendPending = 1
	SMSSendSuccess = 2
	SMSSendFailed  = 3
)

// 微信应用类型，枚举从 1 开始。
const (
	WechatOfficial = 1
	WechatOpen     = 2
	WechatMiniApp  = 3
)

// 支付渠道，枚举从 1 开始。
const (
	PaymentWechat = 1
	PaymentAlipay = 2
)
