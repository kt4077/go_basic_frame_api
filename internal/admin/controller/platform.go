package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
)

// PlatformController 平台配置控制器。
type PlatformController struct{ Logic *logic.PlatformLogic }

// AdminDetail 查询管理端配置。
func (h *PlatformController) AdminDetail(c *gin.Context) {
	result, err := h.Logic.AdminDetail(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}

// SaveAdmin 保存管理端配置。
func (h *PlatformController) SaveAdmin(c *gin.Context) {
	var req param.AdminPlatformSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	result, err := h.Logic.SaveAdmin(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}

// UserDetail 查询用户端配置。
func (h *PlatformController) UserDetail(c *gin.Context) {
	result, err := h.Logic.UserDetail(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}

// SaveUser 保存用户端配置。
func (h *PlatformController) SaveUser(c *gin.Context) {
	var req param.UserPlatformSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	result, err := h.Logic.SaveUser(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
