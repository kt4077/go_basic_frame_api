package controller

import (
	"github.com/gin-gonic/gin"
	"server_api/internal/plugins/news/admin/logic"
	"server_api/internal/plugins/news/admin/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

// ArticleController 新闻文章管理控制器。
type ArticleController struct{ Logic *logic.ArticleLogic }

func (h *ArticleController) List(c *gin.Context) {
	var req param.ArticleListReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.List(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ArticleController) Detail(c *gin.Context) {
	var req param.ArticleIDReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.Detail(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ArticleController) Save(c *gin.Context) {
	var req param.ArticleSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.Save(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *ArticleController) UpdateStatus(c *gin.Context) {
	var req param.ArticleStatusReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if err := h.Logic.UpdateStatus(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
func (h *ArticleController) Delete(c *gin.Context) {
	var req param.ArticleIDReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if err := h.Logic.Delete(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
