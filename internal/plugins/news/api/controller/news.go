package controller

import (
	"github.com/gin-gonic/gin"
	"server_api/internal/plugins/news/api/logic"
	"server_api/internal/plugins/news/api/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

// NewsController 用户端新闻控制器。
type NewsController struct{ Logic *logic.NewsLogic }

func (h *NewsController) Categories(c *gin.Context) {
	result, err := h.Logic.Categories(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *NewsController) Articles(c *gin.Context) {
	var req param.ArticleListReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.Articles(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *NewsController) Article(c *gin.Context) {
	var req param.ArticleDetailReq
	if err := requestvalidate.BindURI(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.Article(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
