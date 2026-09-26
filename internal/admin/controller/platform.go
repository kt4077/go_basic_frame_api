package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
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
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
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
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.SaveUser(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
