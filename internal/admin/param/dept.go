package param

type DeptSaveReq struct {
	ID       uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name     string `json:"name" binding:"required" validate:"部门名称" comment:"部门名称"`
	ParentID uint   `json:"parent_id" validate:"上级部门ID" comment:"上级部门ID"`
	Sort     int    `json:"sort" validate:"排序" comment:"排序"`
	Leader   string `json:"leader" validate:"负责人" comment:"负责人"`
	Remark   string `json:"remark" validate:"备注" comment:"备注"`
}
