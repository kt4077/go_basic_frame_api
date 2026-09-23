package param

type MenuListReq struct {
	Name string `form:"name" comment:"菜单名称"`
	Type int    `form:"type" comment:"菜单类型"`
}

type MenuSaveReq struct {
	ID       uint   `json:"id" comment:"主键ID"`
	Name     string `json:"name" binding:"required" comment:"菜单名称"`
	Type     int    `json:"type" binding:"required,oneof=1 2 3" comment:"菜单类型"`
	ParentID uint   `json:"parent_id" comment:"上级菜单ID"`
	Path     string `json:"path" comment:"前端路由地址"`
	ApiPath  string `json:"api_path" comment:"接口权限地址"`
	Icon     string `json:"icon" comment:"菜单图标"`
	Sort     int    `json:"sort" comment:"排序"`
	Status   int    `json:"status" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}
