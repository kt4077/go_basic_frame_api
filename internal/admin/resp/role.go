package resp

type RoleListRes struct {
	List  []RoleItem `json:"list" comment:"角色列表"`
	Total int64      `json:"total" comment:"数据总数"`
}
