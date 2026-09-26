package param

type RoleListReq struct {
	Name     string `form:"name" validate:"角色名称" comment:"角色名称"`
	Page     int    `form:"page" validate:"页码" comment:"页码"`
	PageSize int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}

type RoleSaveReq struct {
	ID       uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name     string `json:"name" binding:"required" validate:"角色名称" comment:"角色名称"`
	Code     string `json:"code" binding:"required" validate:"角色编码" comment:"角色编码"`
	ParentID uint   `json:"parent_id" validate:"上级角色ID" comment:"上级角色ID"`
	Sort     int    `json:"sort" validate:"排序" comment:"排序"`
	Status   int    `json:"status" validate:"状态" comment:"状态"`
	Remark   string `json:"remark" validate:"备注" comment:"备注"`
}

type AssignMenusReq struct {
	RoleID  uint   `json:"role_id" binding:"required" validate:"角色ID" comment:"角色ID"`
	MenuIDs []uint `json:"menu_ids" validate:"菜单ID列表" comment:"菜单ID列表"`
}
