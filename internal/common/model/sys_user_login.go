package model

import "time"

// SysUserLogin 登录流水表。每次登录生成一条记录，
// login_id 写入 JWT，同时作为 Redis 缓存键，用于主动过期、禁用、踢人下线。
// 状态枚举见 internal/common/enums/login.go。表：sys_user_login
type SysUserLogin struct {
	ID        uint       `gorm:"primaryKey;comment:主键ID" json:"id"`
	LoginID   string     `gorm:"size:64;not null;uniqueIndex;comment:登录唯一标识" json:"login_id"`
	UserID    uint       `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Username  string     `gorm:"size:32;comment:登录账号" json:"username"`
	Client    string     `gorm:"size:16;default:'';comment:客户端类型" json:"client"`
	LoginIP   string     `gorm:"size:64;default:'';comment:登录IP" json:"login_ip"`
	UserAgent string     `gorm:"size:255;default:'';comment:用户代理" json:"user_agent"`
	LoginAt   time.Time  `gorm:"comment:登录时间" json:"login_at"`
	LogoutAt  *time.Time `gorm:"comment:退出时间" json:"logout_at"`
	Status    int        `gorm:"default:1;comment:登录状态" json:"status"`
}

func (SysUserLogin) TableName() string { return "sys_user_login" }
