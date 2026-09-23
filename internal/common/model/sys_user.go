package model

// SysUser 账号表，管理端/接口端共用。表：sys_user
type SysUser struct {
	Base
	Username string `gorm:"size:32;not null;uniqueIndex;comment:登录账号" json:"username"`
	Password string `gorm:"size:128;not null;comment:登录密码密文" json:"-"`
	Nickname string `gorm:"size:32;default:'';comment:用户昵称" json:"nickname"`
	Avatar   string `gorm:"size:512;default:'';comment:头像相对路径" json:"avatar"`
	Mobile   string `gorm:"size:16;default:'';comment:手机号" json:"mobile"`
	Email    string `gorm:"size:64;default:'';comment:邮箱" json:"email"`
	DeptID   uint   `gorm:"default:0;index;comment:所属部门ID" json:"dept_id"`
	Status   int    `gorm:"default:1;index;comment:状态" json:"status"`
	IsSuper  int    `gorm:"default:0;comment:是否超级管理员" json:"is_super"`
}

func (SysUser) TableName() string { return "sys_user" }
