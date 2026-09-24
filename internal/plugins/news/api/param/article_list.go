package param

// ArticleListReq 用户端新闻文章列表请求。
type ArticleListReq struct {
	CategorySlug string `form:"category_slug" binding:"omitempty,max=64" comment:"分类标识"`
	Keyword      string `form:"keyword" binding:"omitempty,max=100" comment:"标题关键字"`
	Page         int    `form:"page" comment:"页码"`
	PageSize     int    `form:"page_size" comment:"每页数量"`
}

// ArticleDetailReq 用户端新闻文章详情请求。
type ArticleDetailReq struct {
	Slug string `uri:"slug" binding:"required,max=128" comment:"文章标识"`
}
