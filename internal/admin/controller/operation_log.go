package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
)

type OperationLogController struct{ Logic *logic.OperationLogLogic }

// List 操作日志分页列表。
func (h *OperationLogController) List(c *gin.Context) {
	var req param.OperationLogListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	res, err := h.Logic.List(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Delete 批量物理删除操作日志。
func (h *OperationLogController) Delete(c *gin.Context) {
	var req param.OperationLogDeleteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "请选择需要删除的操作日志，单次最多500条")
		return
	}
	res, err := h.Logic.Delete(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Clear 全量物理清空操作日志。
func (h *OperationLogController) Clear(c *gin.Context) {
	var req param.OperationLogClearReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "清空确认信息错误")
		return
	}
	res, err := h.Logic.Clear(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}
