package model

import "time"

// SysPluginInstallLog 记录插件安装和升级结果。表：sys_plugin_install_log
type SysPluginInstallLog struct {
	ID           uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	PluginID     string    `gorm:"size:64;not null;index;comment:插件唯一标识" json:"plugin_id"`
	Version      string    `gorm:"size:32;not null;comment:目标版本" json:"version"`
	Action       int       `gorm:"not null;comment:操作类型，1安装，2升级" json:"action"`
	Status       int       `gorm:"not null;comment:执行状态，1成功，2失败" json:"status"`
	PackageHash  string    `gorm:"size:64;not null;default:'';comment:发行包SHA256" json:"package_hash"`
	ErrorMessage string    `gorm:"size:1000;not null;default:'';comment:失败原因" json:"error_message"`
}

// TableName 返回插件安装日志表名。
func (SysPluginInstallLog) TableName() string { return "sys_plugin_install_log" }
