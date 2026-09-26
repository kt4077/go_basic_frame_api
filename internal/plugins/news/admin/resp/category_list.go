package resp

import "time"

// CategoryItem 新闻分类响应结构。
type CategoryItem struct {
	ID        uint      `json:"id" gorm:"column:id" comment:"主键ID"`
	Name      string    `json:"name" gorm:"column:name" comment:"分类名称"`
	Slug      string    `json:"slug" gorm:"column:slug" comment:"分类标识"`
	Sort      int       `json:"sort" gorm:"column:sort" comment:"排序值"`
	Status    int       `json:"status" gorm:"column:status" comment:"状态，1启用，2停用"`
	Remark    string    `json:"remark" gorm:"column:remark" comment:"分类备注"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at" comment:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at" comment:"更新时间"`
}

// TableName 返回分类响应查询表名。
func (CategoryItem) TableName() string { return "plg_news_category" }

// CategoryListRes 新闻分类分页响应。
type CategoryListRes struct {
	List  []CategoryItem `json:"list" comment:"分类列表"`
	Total int64          `json:"total" comment:"数据总数"`
}
