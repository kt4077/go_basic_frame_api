package param

// PluginListReq 插件分页查询参数。
type PluginListReq struct {
	Keyword  string `form:"keyword" validate:"插件名称或标识关键字" comment:"插件名称或标识关键字"`
	Status   int    `form:"status" binding:"omitempty,oneof=1 2" validate:"插件状态" comment:"插件状态，1启用，2停用"`
	Page     int    `form:"page" validate:"页码" comment:"页码"`
	PageSize int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}

// PluginIDReq 插件标识参数。
type PluginIDReq struct {
	PluginID string `json:"plugin_id" form:"plugin_id" binding:"required" validate:"插件唯一标识" comment:"插件唯一标识"`
}

// PluginStatusReq 插件状态修改参数。
type PluginStatusReq struct {
	PluginID string `json:"plugin_id" binding:"required" validate:"插件唯一标识" comment:"插件唯一标识"`
	Status   int    `json:"status" binding:"required,oneof=1 2" validate:"插件状态" comment:"插件状态，1启用，2停用"`
}

// PluginInfoUpdateReq 插件展示信息修改参数。
type PluginInfoUpdateReq struct {
	PluginID    string `json:"plugin_id" binding:"required" validate:"插件唯一标识" comment:"插件唯一标识"`
	Logo        string `json:"logo" binding:"max=1000" validate:"插件Logo相对路径" comment:"插件Logo相对路径"`
	Author      string `json:"author" binding:"max=100" validate:"插件作者" comment:"插件作者"`
	Homepage    string `json:"homepage" binding:"omitempty,url,max=500" validate:"插件主页地址" comment:"插件主页地址"`
	Description string `json:"description" binding:"max=2000" validate:"插件描述" comment:"插件描述"`
}
