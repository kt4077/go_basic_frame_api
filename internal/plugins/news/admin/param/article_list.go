package param

// ArticleListReq 新闻文章列表请求。
type ArticleListReq struct {
	Keyword      string `form:"keyword" binding:"omitempty,max=200" validate:"标题或标识关键字" comment:"标题或标识关键字"`
	CategoryID   uint   `form:"category_id" validate:"分类ID" comment:"分类ID"`
	Status       int    `form:"status" binding:"omitempty,oneof=1 2 3" validate:"状态" comment:"状态，1草稿，2已发布，3已下线"`
	EnableStatus int    `form:"enable_status" binding:"omitempty,oneof=1 2" validate:"启用状态" comment:"启用状态，1启用，2禁用"`
	Page         int    `form:"page" validate:"页码" comment:"页码"`
	PageSize     int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}
