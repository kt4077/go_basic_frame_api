package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
)

type MemberController struct{ Logic *logic.MemberLogic }

// List 会员用户分页列表。
func (h *MemberController) List(c *gin.Context) {
	var req param.MemberListReq
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

// SetStatus 启用/禁用会员账号。
func (h *MemberController) SetStatus(c *gin.Context) {
	var req param.MemberSetStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	if err := h.Logic.SetStatus(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
