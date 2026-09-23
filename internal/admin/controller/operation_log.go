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
