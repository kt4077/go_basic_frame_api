package resp

import (
	"time"

	"server_api/internal/common/model"
	"server_api/pkg/mask"
)

type MemberListRes struct {
	List  []MemberItem `json:"list" comment:"用户列表"`
	Total int64        `json:"total" comment:"数据总数"`
}

type MemberItem struct {
	BaseItem       `comment:"基础字段"`
	SN             string     `json:"sn" comment:"用户编号"`
	Nickname       string     `json:"nickname" comment:"昵称"`
	RealName       string     `json:"real_name" comment:"姓名"`
	Account        string     `json:"account" comment:"登录账号"`
	Mobile         string     `json:"mobile" comment:"手机号（脱敏展示，如138****8001）"`
	Avatar         string     `json:"avatar" comment:"头像地址"`
	Gender         int        `json:"gender" comment:"性别：1男，2女，3未知"`
	Age            int        `json:"age" comment:"年龄"`
	Birthday       string     `json:"birthday" comment:"出生日期，格式YYYY-MM-DD"`
	RegisterIP     string     `json:"register_ip" comment:"注册IP"`
	LoginIP        string     `json:"login_ip" comment:"最近登录IP"`
	RegisteredAt   *time.Time `json:"registered_at" comment:"注册时间"`
	LoggedAt       *time.Time `json:"logged_at" comment:"最近登录时间"`
	Balance        string     `json:"balance" comment:"账户余额"`
	RegisterSource int        `json:"register_source" comment:"注册来源：1微信小程序，2微信公众号，3iOS，4Android"`
	Status         int        `json:"status" comment:"账号状态：1启用，2禁用"`
}

func NewMemberItem(v model.SysMember) MemberItem {
	return MemberItem{
		BaseItem:       base(v.Base),
		SN:             v.SN,
		Nickname:       v.Nickname,
		RealName:       v.RealName,
		Account:        memberAccount(v.Account),
		Mobile:         mask.Mobile(v.Mobile),
		Avatar:         v.Avatar,
		Gender:         v.Gender,
		Age:            v.Age,
		Birthday:       formatMemberDate(v.Birthday),
		RegisterIP:     v.RegisterIP,
		LoginIP:        v.LoginIP,
		RegisteredAt:   v.RegisteredAt,
		LoggedAt:       v.LoggedAt,
		Balance:        v.Balance,
		RegisterSource: v.RegisterSource,
		Status:         v.Status,
	}
}

func NewMemberItems(list []model.SysMember) []MemberItem {
	result := make([]MemberItem, 0, len(list))
	for _, v := range list {
		result = append(result, NewMemberItem(v))
	}
	return result
}

// memberAccount 登录账号为 NULL（未设置）时返回空串。
func memberAccount(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// formatMemberDate 出生日期只保留日期部分，未设置返回空串。
func formatMemberDate(v *time.Time) string {
	if v == nil {
		return ""
	}
	return v.Format("2006-01-02")
}
