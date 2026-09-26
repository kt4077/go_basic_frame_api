package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

type DeptController struct{ Logic *logic.DeptLogic }

// Tree 部门树。
func (h *DeptController) Tree(c *gin.Context) {
	res, err := h.Logic.Tree(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Create 新增部门。
func (h *DeptController) Create(c *gin.Context) {
	var req param.DeptSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	res, err := h.Logic.Create(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Update 修改部门。
func (h *DeptController) Update(c *gin.Context) {
	var req param.DeptSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	res, err := h.Logic.Update(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Delete 删除部门。
func (h *DeptController) Delete(c *gin.Context) {
	var req param.IDReq
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
