package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/pkg/response"
)

type DashboardController struct{ Logic *logic.DashboardLogic }

// Overview 系统总览数据。
func (h *DashboardController) Overview(c *gin.Context) {
	res, err := h.Logic.Overview(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}
