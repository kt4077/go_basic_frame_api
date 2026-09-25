// Package param 用户端请求参数定义，按业务功能组织文件。
package param

// SmsCodeReq 发送短信验证码。
type SmsCodeReq struct {
	Mobile string `json:"mobile" binding:"required" comment:"手机号"`
	Scene  int    `json:"scene" binding:"required,oneof=1 2 3" comment:"验证码场景：1登录，2注册，3换绑手机号"`
}
