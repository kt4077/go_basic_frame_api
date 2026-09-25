// Package resp 用户端响应参数定义，按业务功能组织文件。
package resp

import "time"

type LoginRes struct {
	Token string `json:"token" comment:"访问令牌"`
}

// MemberProfile 当前会员资料。
type MemberProfile struct {
	ID             uint      `json:"id" comment:"用户ID"`
	SN             string    `json:"sn" comment:"用户编号"`
	Account        string    `json:"account" comment:"登录账号"`
	Nickname       string    `json:"nickname" comment:"昵称"`
	RealName       string    `json:"real_name" comment:"姓名"`
	Avatar         string    `json:"avatar" comment:"头像地址"`
	Mobile         string    `json:"mobile" comment:"手机号（脱敏展示，如138****8001）"`
	Gender         int       `json:"gender" comment:"性别：1男，2女，3未知"`
	RegisterSource int       `json:"register_source" comment:"注册来源：1微信小程序，2微信公众号，3iOS，4Android"`
	RegisteredAt   time.Time `json:"registered_at" comment:"注册时间"`
}
