// Package controller 管理端控制器：只做参数绑定与响应，业务逻辑调用 logic 层。
package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
)

type AuthController struct{ Logic *logic.AuthLogic }

// Login 管理端登录。
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

// Me 当前登录用户信息。
func (h *AuthController) Me(c *gin.Context) {
	res, err := h.Logic.Me(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// UpdateProfile 修改当前管理员个人资料。
func (h *AuthController) UpdateProfile(c *gin.Context) {
	var req param.ProfileUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "个人资料参数错误")
		return
	}
	res, err := h.Logic.UpdateProfile(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// UpdateAvatar 修改当前管理员头像。
func (h *AuthController) UpdateAvatar(c *gin.Context) {
	var req param.AvatarUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeErrParams, "头像参数错误")
		return
	}
	res, err := h.Logic.UpdateAvatar(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// GetRouters 当前用户的菜单树。
func (h *AuthController) GetRouters(c *gin.Context) {
	res, err := h.Logic.GetRouters(c)
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

// Permissions 当前用户的接口权限列表（前端按钮显隐用）。
func (h *AuthController) Permissions(c *gin.Context) {
	res, err := h.Logic.Permissions(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}
