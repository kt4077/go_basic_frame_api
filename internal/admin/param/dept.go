package param

type DeptSaveReq struct {
	ID       uint   `json:"id" comment:"主键ID"`
	Name     string `json:"name" binding:"required" comment:"部门名称"`
	ParentID uint   `json:"parent_id" comment:"上级部门ID"`
	Sort     int    `json:"sort" comment:"排序"`
	Leader   string `json:"leader" comment:"负责人"`
	Remark   string `json:"remark" comment:"备注"`
}
