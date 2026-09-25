// Package controller 用户端接口层。只负责参数绑定、调用 logic 和统一响应。
package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/api/logic"
	"server_api/internal/api/param"
	"server_api/pkg/response"
)

type AuthController struct{ Logic *logic.AuthLogic }

// Register 手机号验证码注册。
func (h *AuthController) Register(c *gin.Context) {
	var req param.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	res, err := h.Logic.Register(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// LoginPassword 账号密码登录。
func (h *AuthController) LoginPassword(c *gin.Context) {
	var req param.PasswordLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	res, err := h.Logic.LoginPassword(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// LoginSms 手机号验证码登录。
func (h *AuthController) LoginSms(c *gin.Context) {
	var req param.SmsLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	res, err := h.Logic.LoginSms(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Logout 退出登录。
func (h *AuthController) Logout(c *gin.Context) {
	if err := h.Logic.Logout(c); err != nil {
		response.Fail(c, response.CodeErrBusiness, "退出失败")
		return
	}
	response.OK(c, nil)
}

// Profile 当前会员资料。
func (h *AuthController) Profile(c *gin.Context) {
	res, err := h.Logic.Profile(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// UpdateProfile 修改昵称、姓名与头像。
func (h *AuthController) UpdateProfile(c *gin.Context) {
	var req param.ProfileUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	res, err := h.Logic.UpdateProfile(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// UpdateAccount 修改登录账号。
func (h *AuthController) UpdateAccount(c *gin.Context) {
	var req param.AccountUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	if err := h.Logic.UpdateAccount(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}

// UpdatePassword 修改登录密码。
func (h *AuthController) UpdatePassword(c *gin.Context) {
	var req param.PasswordUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	if err := h.Logic.UpdatePassword(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}

// UpdateMobile 修改手机号。
func (h *AuthController) UpdateMobile(c *gin.Context) {
	var req param.MobileUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误")
		return
	}
	if err := h.Logic.UpdateMobile(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
