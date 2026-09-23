// Package resp 管理端响应参数定义，与 controller 分离并按业务功能组织文件。
package resp

type LoginRes struct {
	Token string `json:"token" comment:"访问令牌"`
}
