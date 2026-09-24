package param

// CategoryListReq 新闻分类列表请求。
type CategoryListReq struct {
	Keyword  string `form:"keyword" binding:"omitempty,max=64" comment:"分类名称或标识关键字"`
	Status   int    `form:"status" binding:"omitempty,oneof=1 2" comment:"状态，1启用，2停用"`
	Page     int    `form:"page" comment:"页码"`
	PageSize int    `form:"page_size" comment:"每页数量"`
}
