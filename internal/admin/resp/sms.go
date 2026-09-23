package resp

type SMSLogListRes struct {
	List  []SMSLogItem `json:"list" comment:"短信发送记录"`
	Total int64        `json:"total" comment:"数据总数"`
}
