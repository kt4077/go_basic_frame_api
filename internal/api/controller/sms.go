// Package controller 用户端接口层。只负责参数绑定、调用 logic 和统一响应。
package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/api/logic"
	"server_api/internal/api/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

type SmsController struct{ Logic *logic.SmsLogic }

// SendCode 发送短信验证码。
func (h *SmsController) SendCode(c *gin.Context) {
	var req param.SmsCodeReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if err := h.Logic.SendSmsCode(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
