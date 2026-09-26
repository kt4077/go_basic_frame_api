package resp

// CategoryItem 用户端新闻分类响应。
type CategoryItem struct {
	ID   uint   `json:"id" gorm:"column:id" comment:"分类ID"`
	Name string `json:"name" gorm:"column:name" comment:"分类名称"`
	Slug string `json:"slug" gorm:"column:slug" comment:"分类标识"`
}

// TableName 返回分类响应查询表名。
func (CategoryItem) TableName() string { return "plg_news_category" }
