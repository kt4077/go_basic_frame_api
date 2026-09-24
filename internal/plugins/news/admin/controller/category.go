package controller

import (
	"github.com/gin-gonic/gin"
	"server_api/internal/plugins/news/admin/logic"
	"server_api/internal/plugins/news/admin/param"
	"server_api/pkg/response"
)

// CategoryController 新闻分类管理控制器。
type CategoryController struct{ Logic *logic.CategoryLogic }

func (h *CategoryController) List(c *gin.Context) {
	var req param.CategoryListReq
	if c.ShouldBindQuery(&req) != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	result, err := h.Logic.List(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *CategoryController) Save(c *gin.Context) {
	var req param.CategorySaveReq
	if c.ShouldBindJSON(&req) != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	result, err := h.Logic.Save(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
func (h *CategoryController) Delete(c *gin.Context) {
	var req param.CategoryIDReq
	if c.ShouldBindJSON(&req) != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	if err := h.Logic.Delete(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
