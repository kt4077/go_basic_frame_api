package param

type LoginReq struct {
	Username string `json:"username" binding:"required" validate:"登录账号" comment:"登录账号"`
	Password string `json:"password" binding:"required" validate:"登录密码" comment:"登录密码"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password" binding:"required" validate:"原密码" comment:"原密码"`
	NewPassword string `json:"new_password" binding:"required,min=6" validate:"新密码" comment:"新密码"`
}
