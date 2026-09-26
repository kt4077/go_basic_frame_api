package param

// ArticleListReq 用户端新闻文章列表请求。
type ArticleListReq struct {
	CategorySlug string `form:"category_slug" binding:"omitempty,max=64" validate:"分类标识" comment:"分类标识"`
	Keyword      string `form:"keyword" binding:"omitempty,max=100" validate:"标题关键字" comment:"标题关键字"`
	Page         int    `form:"page" validate:"页码" comment:"页码"`
	PageSize     int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}

// ArticleDetailReq 用户端新闻文章详情请求。
type ArticleDetailReq struct {
	Slug string `uri:"slug" binding:"required,max=128" validate:"文章标识" comment:"文章标识"`
}
