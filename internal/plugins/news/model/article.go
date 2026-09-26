package model

import (
	"time"

	"gorm.io/gorm"
)

// Article 新闻文章数据库模型，不在Model中定义关联关系。表：plg_news_article。
type Article struct {
	ID               uint           `gorm:"primaryKey;comment:主键ID" json:"id"`
	CategoryID       uint           `gorm:"not null;index:idx_plg_news_article_category_status;comment:分类ID" json:"category_id"`
	Title            string         `gorm:"size:200;not null;comment:文章标题" json:"title"`
	Slug             string         `gorm:"size:128;not null;uniqueIndex:uk_plg_news_article_slug;comment:文章标识" json:"slug"`
	Summary          string         `gorm:"size:500;not null;default:'';comment:文章摘要" json:"summary"`
	Cover            string         `gorm:"size:1024;not null;default:'';comment:封面相对路径" json:"cover"`
	Content          string         `gorm:"type:longtext;not null;comment:文章正文" json:"content"`
	Status           int            `gorm:"not null;default:1;index:idx_plg_news_article_category_status;comment:状态，1草稿，2已发布，3已下线" json:"status"`
	EnableStatus     int            `gorm:"not null;default:1;comment:启用状态，1启用，2禁用" json:"enable_status"`
	Sort             int            `gorm:"not null;default:0;index;comment:排序值" json:"sort"`
	ViewCount        uint64         `gorm:"not null;default:0;comment:真实浏览次数" json:"view_count"`
	VirtualViewCount uint64         `gorm:"not null;default:0;comment:虚拟浏览次数" json:"virtual_view_count"`
	PublishedAt      *time.Time     `gorm:"index;comment:发布时间" json:"published_at"`
	CreatedAt        time.Time      `gorm:"comment:创建时间" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"comment:更新时间" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;comment:删除时间" json:"-"`
}

// TableName 返回新闻文章表名。
func (Article) TableName() string { return "plg_news_article" }
