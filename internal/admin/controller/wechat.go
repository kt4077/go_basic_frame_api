package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
)

type WechatController struct{ Logic *logic.WechatLogic }

func (h *WechatController) List(c *gin.Context) {
	configType, _ := strconv.Atoi(c.Query("type"))
	result, err := h.Logic.List(c, configType)
	respond(c, result, err)
}

func (h *WechatController) Save(c *gin.Context) {
	var req param.WechatConfigSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	result, err := h.Logic.Save(c, &req)
	respond(c, result, err)
}

func (h *WechatController) Delete(c *gin.Context) {
	var req param.IDReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误，缺少ID")
		return
	}
	if err := h.Logic.Delete(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
