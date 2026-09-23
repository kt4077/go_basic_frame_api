package resp

type UserListRes struct {
	List  []UserItem `json:"list" comment:"用户列表"`
	Total int64      `json:"total" comment:"数据总数"`
}
