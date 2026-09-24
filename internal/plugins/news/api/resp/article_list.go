package resp

import "time"

// ArticleCategoryItem 文章关联分类响应，关联只在Resp定义。
type ArticleCategoryItem struct {
	ID   uint   `json:"id" gorm:"column:id" comment:"分类ID"`
	Name string `json:"name" gorm:"column:name" comment:"分类名称"`
	Slug string `json:"slug" gorm:"column:slug" comment:"分类标识"`
}

func (ArticleCategoryItem) TableName() string { return "plg_news_category" }

// ArticleItem 用户端新闻文章响应。
type ArticleItem struct {
	ID               uint                `json:"id" gorm:"column:id" comment:"文章ID"`
	CategoryID       uint                `json:"category_id" gorm:"column:category_id" comment:"分类ID"`
	Category         ArticleCategoryItem `json:"category" gorm:"foreignKey:CategoryID;references:ID" comment:"关联分类"`
	Title            string              `json:"title" gorm:"column:title" comment:"文章标题"`
	Slug             string              `json:"slug" gorm:"column:slug" comment:"文章标识"`
	Summary          string              `json:"summary" gorm:"column:summary" comment:"文章摘要"`
	Cover            string              `json:"cover" gorm:"column:cover" comment:"封面相对路径"`
	CoverURL         string              `json:"cover_url" gorm:"-" comment:"封面完整访问地址"`
	Content          string              `json:"content,omitempty" gorm:"column:content" comment:"文章正文"`
	ViewCount        uint64              `json:"view_count" gorm:"column:view_count" comment:"真实浏览次数"`
	VirtualViewCount uint64              `json:"virtual_view_count" gorm:"column:virtual_view_count" comment:"虚拟浏览次数"`
	TotalViewCount   uint64              `json:"total_view_count" gorm:"-" comment:"展示浏览次数，真实与虚拟之和"`
	PublishedAt      *time.Time          `json:"published_at" gorm:"column:published_at" comment:"发布时间"`
}

func (ArticleItem) TableName() string { return "plg_news_article" }

// ArticleListRes 用户端新闻文章分页响应。
type ArticleListRes struct {
	List  []ArticleItem `json:"list" comment:"文章列表"`
	Total int64         `json:"total" comment:"数据总数"`
}
