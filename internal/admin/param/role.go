package param

type RoleListReq struct {
	Name     string `form:"name" comment:"角色名称"`
	Page     int    `form:"page" comment:"页码"`
	PageSize int    `form:"page_size" comment:"每页数量"`
}

type RoleSaveReq struct {
	ID       uint   `json:"id" comment:"主键ID"`
	Name     string `json:"name" binding:"required" comment:"角色名称"`
	Code     string `json:"code" binding:"required" comment:"角色编码"`
	ParentID uint   `json:"parent_id" comment:"上级角色ID"`
	Sort     int    `json:"sort" comment:"排序"`
	Status   int    `json:"status" comment:"状态"`
	Remark   string `json:"remark" comment:"备注"`
}

type AssignMenusReq struct {
	RoleID  uint   `json:"role_id" binding:"required" comment:"角色ID"`
	MenuIDs []uint `json:"menu_ids" comment:"菜单ID列表"`
}
