package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

type SMSController struct{ Logic *logic.SMSLogic }

func (h *SMSController) ConfigList(c *gin.Context) {
	result, err := h.Logic.ConfigList(c)
	respond(c, result, err)
}
func (h *SMSController) SignatureList(c *gin.Context) {
	result, err := h.Logic.SignatureList(c)
	respond(c, result, err)
}
func (h *SMSController) TemplateList(c *gin.Context) {
	result, err := h.Logic.TemplateList(c)
	respond(c, result, err)
}

func (h *SMSController) SaveConfig(c *gin.Context) {
	var req param.SMSConfigSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.SaveConfig(c, &req)
	respond(c, result, err)
}

func (h *SMSController) SaveSignature(c *gin.Context) {
	var req param.SMSSignatureSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.SaveSignature(c, &req)
	respond(c, result, err)
}

func (h *SMSController) SaveTemplate(c *gin.Context) {
	var req param.SMSTemplateSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.SaveTemplate(c, &req)
	respond(c, result, err)
}

func (h *SMSController) DeleteConfig(c *gin.Context)    { h.delete(c, h.Logic.DeleteConfig) }
func (h *SMSController) DeleteSignature(c *gin.Context) { h.delete(c, h.Logic.DeleteSignature) }
func (h *SMSController) DeleteTemplate(c *gin.Context)  { h.delete(c, h.Logic.DeleteTemplate) }

func (h *SMSController) delete(c *gin.Context, fn func(*gin.Context, *param.IDReq) error) {
	var req param.IDReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if err := fn(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}

func (h *SMSController) LogList(c *gin.Context) {
	var req param.SMSLogListReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.LogList(c, &req)
	respond(c, result, err)
}

func respond(c *gin.Context, result interface{}, err error) {
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
