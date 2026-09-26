package resp

type SMSLogListRes struct {
	List  []SMSLogItem `json:"list" comment:"短信发送记录"`
	Total int64        `json:"total" comment:"数据总数"`
}

// SMSConfigTestRes 短信渠道测试结果。
type SMSConfigTestRes struct {
	ProviderMessageID string `json:"provider_message_id" comment:"短信服务商消息ID"`
}
