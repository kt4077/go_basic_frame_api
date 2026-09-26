package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

type PaymentController struct{ Logic *logic.PaymentLogic }

func (h *PaymentController) List(c *gin.Context) {
	var req param.PaymentConfigListReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.List(c, req.Channel)
	respond(c, result, err)
}

func (h *PaymentController) Save(c *gin.Context) {
	var req param.PaymentConfigSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	result, err := h.Logic.Save(c, &req)
	respond(c, result, err)
}

func (h *PaymentController) Delete(c *gin.Context) {
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
