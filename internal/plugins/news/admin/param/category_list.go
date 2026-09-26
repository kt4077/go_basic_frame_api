package param

// CategoryListReq 新闻分类列表请求。
type CategoryListReq struct {
	Keyword  string `form:"keyword" binding:"omitempty,max=64" validate:"分类名称或标识关键字" comment:"分类名称或标识关键字"`
	Status   int    `form:"status" binding:"omitempty,oneof=1 2" validate:"状态" comment:"状态，1启用，2停用"`
	Page     int    `form:"page" validate:"页码" comment:"页码"`
	PageSize int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}
