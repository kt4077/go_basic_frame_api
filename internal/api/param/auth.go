// Package param 用户端请求参数定义，按业务功能组织文件。
package param

// RegisterReq 手机号验证码注册。
type RegisterReq struct {
	Mobile string `json:"mobile" binding:"required" validate:"手机号" comment:"手机号"`
	Code   string `json:"code" binding:"required,len=6" validate:"短信验证码" comment:"短信验证码"`
}

// PasswordLoginReq 账号密码登录。
type PasswordLoginReq struct {
	Account  string `json:"account" binding:"required" validate:"登录账号" comment:"登录账号"`
	Password string `json:"password" binding:"required" validate:"登录密码" comment:"登录密码"`
}

// SmsLoginReq 手机号验证码登录。
type SmsLoginReq struct {
	Mobile string `json:"mobile" binding:"required" validate:"手机号" comment:"手机号"`
	Code   string `json:"code" binding:"required,len=6" validate:"短信验证码" comment:"短信验证码"`
}

// ProfileUpdateReq 会员资料修改（昵称、姓名、头像、性别）。
type ProfileUpdateReq struct {
	Nickname string `json:"nickname" binding:"omitempty,max=32" validate:"昵称" comment:"昵称"`
	RealName string `json:"real_name" binding:"omitempty,max=32" validate:"姓名" comment:"姓名"`
	Avatar   string `json:"avatar" binding:"omitempty,max=1024" validate:"头像地址" comment:"头像地址"`
	Gender   int    `json:"gender" binding:"omitempty" validate:"性别" comment:"性别：1男，2女，3未知；不传表示不修改"`
}

// AccountUpdateReq 登录账号修改。
type AccountUpdateReq struct {
	Account string `json:"account" binding:"required,min=4,max=32" validate:"登录账号" comment:"登录账号"`
}

// PasswordUpdateReq 登录密码修改。
type PasswordUpdateReq struct {
	OldPassword string `json:"old_password" binding:"omitempty" validate:"原密码" comment:"原密码，已设置密码时必填"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=32" validate:"新密码" comment:"新密码"`
}

// MobileUpdateReq 手机号修改（需新手机号短信验证码）。
type MobileUpdateReq struct {
	Mobile string `json:"mobile" binding:"required" validate:"新手机号" comment:"新手机号"`
	Code   string `json:"code" binding:"required,len=6" validate:"短信验证码" comment:"短信验证码"`
}
