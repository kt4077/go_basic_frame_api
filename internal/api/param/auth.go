// Package param 用户端请求参数定义，按业务功能组织文件。
package param

type LoginReq struct {
	Username string `json:"username" binding:"required" comment:"登录账号"`
	Password string `json:"password" binding:"required" comment:"登录密码"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required" comment:"原密码"`
	NewPassword string `json:"new_password" binding:"required,min=6" comment:"新密码"`
}
