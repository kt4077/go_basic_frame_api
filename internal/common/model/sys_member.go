package model

import "time"

// SysMember 会员（C端用户）账号表。表：sys_member
type SysMember struct {
	Base
	SN             string     `gorm:"column:sn;size:32;not null;uniqueIndex;comment:用户编号，全局唯一" json:"sn"`
	Nickname       string     `gorm:"size:32;default:'';comment:昵称" json:"nickname"`
	RealName       string     `gorm:"size:32;default:'';comment:姓名" json:"real_name"`
	Account        *string    `gorm:"column:account;size:32;uniqueIndex;comment:登录账号，唯一，未设置为NULL" json:"account"`
	Mobile         string     `gorm:"size:16;not null;uniqueIndex;comment:手机号" json:"mobile"`
	Password       string     `gorm:"size:128;default:'';comment:登录密码密文" json:"-"`
	LoginID        string     `gorm:"column:login_id;size:64;default:'';comment:当前登录会话ID，重新登录后旧会话失效" json:"-"`
	Avatar         string     `gorm:"size:512;default:'';comment:头像相对路径" json:"avatar"`
	Gender         int        `gorm:"default:3;comment:性别：1男，2女，3未知" json:"gender"`
	Age            int        `gorm:"default:0;comment:年龄" json:"age"`
	Birthday       *time.Time `gorm:"comment:出生日期" json:"birthday"`
	RegisterIP     string     `gorm:"size:64;default:'';comment:注册IP" json:"register_ip"`
	LoginIP        string     `gorm:"size:64;default:'';comment:最近登录IP" json:"login_ip"`
	RegisteredAt   *time.Time `gorm:"comment:注册时间" json:"registered_at"`
	LoggedAt       *time.Time `gorm:"comment:最近登录时间" json:"logged_at"`
	Balance        string     `gorm:"type:decimal(12,2);not null;default:'0.00';comment:账户余额" json:"balance"`
	RegisterSource int        `gorm:"default:1;comment:注册来源：1微信小程序，2微信公众号，3iOS，4Android" json:"register_source"`
	Status         int        `gorm:"default:1;index;comment:账号状态：1启用，2禁用" json:"status"`
}

func (SysMember) TableName() string { return "sys_member" }
