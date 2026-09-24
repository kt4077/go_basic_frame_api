package model

import "time"

// SysPluginMigration 插件数据库迁移执行记录。插件迁移只允许向前执行，
// PluginID 与 Version 组成唯一约束。表：sys_plugin_migration
type SysPluginMigration struct {
	ID           uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	CreatedAt    time.Time `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"comment:更新时间" json:"updated_at"`
	PluginID     string    `gorm:"size:64;not null;uniqueIndex:uk_plugin_migration;comment:插件唯一标识" json:"plugin_id"`
	Version      string    `gorm:"size:32;not null;uniqueIndex:uk_plugin_migration;comment:迁移版本" json:"version"`
	Checksum     string    `gorm:"size:64;not null;default:'';comment:迁移内容SHA256摘要" json:"checksum"`
	Status       int       `gorm:"not null;default:1;index;comment:执行状态，1成功，2失败" json:"status"`
	ExecutionMs  int64     `gorm:"not null;default:0;comment:执行耗时毫秒" json:"execution_ms"`
	ErrorMessage string    `gorm:"size:1000;not null;default:'';comment:失败原因" json:"error_message"`
	ExecutedAt   time.Time `gorm:"not null;comment:执行时间" json:"executed_at"`
}

// TableName 返回插件迁移记录表名。
func (SysPluginMigration) TableName() string { return "sys_plugin_migration" }
