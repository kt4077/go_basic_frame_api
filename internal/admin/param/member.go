package param

type MemberListReq struct {
	Keyword        string `form:"keyword" comment:"搜索关键字（昵称/姓名/账号/手机号）"`
	Status         int    `form:"status" comment:"账号状态：1启用，2禁用"`
	RegisterSource int    `form:"register_source" comment:"注册来源：1微信小程序，2微信公众号，3iOS，4Android"`
	Page           int    `form:"page" comment:"页码"`
	PageSize       int    `form:"page_size" comment:"每页数量"`
}

type MemberSetStatusReq struct {
	ID     uint `json:"id" binding:"required" comment:"用户ID"`
	Status int  `json:"status" binding:"required,oneof=1 2" comment:"目标状态：1启用，2禁用"`
}
