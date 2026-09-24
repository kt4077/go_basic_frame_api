package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
)

// PluginController 插件管理控制器。
type PluginController struct{ Logic *logic.PluginLogic }

// List 插件分页列表。
func (h *PluginController) List(c *gin.Context) {
	var req param.PluginListReq
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

// Detail 插件详情和迁移记录。
func (h *PluginController) Detail(c *gin.Context) {
	var req param.PluginIDReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误，缺少插件标识")
		return
	}
	res, err := h.Logic.Detail(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// UpdateStatus 启用或停用插件，修改后需重启服务生效。
func (h *PluginController) UpdateStatus(c *gin.Context) {
	var req param.PluginStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	if err := h.Logic.UpdateStatus(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}

// UpdateInfo 修改插件作者、主页和描述。
func (h *PluginController) UpdateInfo(c *gin.Context) {
	var req param.PluginInfoUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误，请检查插件地址格式和字段长度")
		return
	}
	if err := h.Logic.UpdateInfo(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
