package param

// ArticleSaveReq 新闻文章保存请求。
type ArticleSaveReq struct {
	ID               uint   `json:"id" comment:"主键ID，新增时为0"`
	CategoryID       uint   `json:"category_id" binding:"required" comment:"分类ID"`
	Title            string `json:"title" binding:"required,max=200" comment:"文章标题"`
	Slug             string `json:"slug" binding:"required,max=128" comment:"文章标识，仅支持字母、数字、中划线和下划线"`
	Summary          string `json:"summary" binding:"max=500" comment:"文章摘要"`
	Cover            string `json:"cover" binding:"omitempty,max=1024" comment:"封面相对路径"`
	Content          string `json:"content" binding:"required,max=2000000" comment:"文章正文，最多200万字符"`
	Sort             int    `json:"sort" binding:"min=0" comment:"排序值"`
	EnableStatus     int    `json:"enable_status" binding:"required,oneof=1 2" comment:"启用状态，1启用，2禁用"`
	VirtualViewCount uint64 `json:"virtual_view_count" comment:"虚拟阅读数量"`
}

// ArticleIDReq 新闻文章ID请求。
type ArticleIDReq struct {
	ID uint `json:"id" form:"id" binding:"required" comment:"文章ID"`
}

// ArticleStatusReq 新闻文章状态变更请求。
type ArticleStatusReq struct {
	ID     uint `json:"id" binding:"required" comment:"文章ID"`
	Status int  `json:"status" binding:"required,oneof=2 3" comment:"目标状态，2发布，3下线"`
}
