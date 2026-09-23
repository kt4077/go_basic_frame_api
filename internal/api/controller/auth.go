// Package controller 用户端控制器：只做参数绑定与响应，业务逻辑调用 logic 层。
package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/api/logic"
	"server_api/internal/api/param"
	"server_api/pkg/response"
)

type AuthController struct{ Logic *logic.AuthLogic }

// Login 用户端登录。
func (h *AuthController) Login(c *gin.Context) {
	var req param.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "用户名和密码不能为空")
		return
	}
	res, err := h.Logic.Login(c, &req)
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

// Profile 当前用户资料。
func (h *AuthController) Profile(c *gin.Context) {
	res, err := h.Logic.Profile(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// ChangePassword 修改自己的密码。
func (h *AuthController) ChangePassword(c *gin.Context) {
	var req param.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "参数错误，新密码至少6位")
		return
	}
	if err := h.Logic.ChangePassword(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
