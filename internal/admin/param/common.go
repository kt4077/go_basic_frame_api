// Package param 管理端请求参数定义，与 controller 分离并按业务功能组织文件。
package param

type IDReq struct {
	ID uint `json:"id" form:"id" binding:"required" validate:"主键ID" comment:"主键ID"`
}
