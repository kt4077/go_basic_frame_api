package param

type MenuListReq struct {
	Name string `form:"name" validate:"菜单名称" comment:"菜单名称"`
	Type int    `form:"type" validate:"菜单类型" comment:"菜单类型"`
}

type MenuSaveReq struct {
	ID       uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name     string `json:"name" binding:"required" validate:"菜单名称" comment:"菜单名称"`
	Type     int    `json:"type" binding:"required,oneof=1 2 3" validate:"菜单类型" comment:"菜单类型"`
	ParentID uint   `json:"parent_id" validate:"上级菜单ID" comment:"上级菜单ID"`
	Path     string `json:"path" validate:"前端路由地址" comment:"前端路由地址"`
	ApiPath  string `json:"api_path" validate:"接口权限地址" comment:"接口权限地址"`
	Icon     string `json:"icon" validate:"菜单图标" comment:"菜单图标"`
	Sort     int    `json:"sort" validate:"排序" comment:"排序"`
	Status   int    `json:"status" validate:"状态" comment:"状态"`
	Remark   string `json:"remark" validate:"备注" comment:"备注"`
}
