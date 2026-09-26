package model

import (
	"time"

	"gorm.io/gorm"
)

// Category 新闻分类数据库模型，不在Model中定义关联关系。表：plg_news_category。
type Category struct {
	ID        uint           `gorm:"primaryKey;comment:主键ID" json:"id"`
	Name      string         `gorm:"size:64;not null;comment:分类名称" json:"name"`
	Slug      string         `gorm:"size:64;not null;uniqueIndex:uk_plg_news_category_slug;comment:分类标识" json:"slug"`
	Sort      int            `gorm:"not null;default:0;index;comment:排序值" json:"sort"`
	Status    int            `gorm:"not null;default:1;index;comment:状态，1启用，2停用" json:"status"`
	Remark    string         `gorm:"size:255;not null;default:'';comment:分类备注" json:"remark"`
	CreatedAt time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 返回新闻分类表名。
func (Category) TableName() string { return "plg_news_category" }
