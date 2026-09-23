// Package model 数据库模型定义。
package model

import (
	"time"

	"gorm.io/gorm"
)

// Base 公共字段。
type Base struct {
	ID        uint           `gorm:"primaryKey;comment:主键ID" json:"id"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}
