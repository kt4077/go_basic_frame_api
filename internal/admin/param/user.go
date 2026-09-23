package param

type UserListReq struct {
	Keyword  string `form:"keyword" comment:"搜索关键字"`
	Status   int    `form:"status" comment:"状态"`
	DeptID   uint   `form:"dept_id" comment:"部门ID"`
	Page     int    `form:"page" comment:"页码"`
	PageSize int    `form:"page_size" comment:"每页数量"`
}

type UserSaveReq struct {
	ID       uint   `json:"id" comment:"主键ID"`
	Username string `json:"username" comment:"登录账号"`
	Password string `json:"password" comment:"登录密码"`
	Nickname string `json:"nickname" comment:"用户昵称"`
	Avatar   string `json:"avatar" binding:"omitempty,max=1024" comment:"头像路径"`
	Mobile   string `json:"mobile" comment:"手机号"`
	Email    string `json:"email" comment:"邮箱"`
	DeptID   uint   `json:"dept_id" comment:"部门ID"`
	Status   int    `json:"status" comment:"状态"`
	IsSuper  int    `json:"is_super" comment:"是否超级管理员"`
	RoleIDs  []uint `json:"role_ids" comment:"角色ID列表"`
}

// ProfileUpdateReq 当前管理员可修改的个人资料。
type ProfileUpdateReq struct {
	Nickname string `json:"nickname" binding:"max=32" comment:"用户昵称"`
	Avatar   string `json:"avatar" binding:"omitempty,max=1024" comment:"头像路径"`
	Mobile   string `json:"mobile" binding:"omitempty,max=16" comment:"手机号"`
	Email    string `json:"email" binding:"omitempty,email,max=64" comment:"邮箱"`
}

// AvatarUpdateReq 当前管理员头像更新参数。
type AvatarUpdateReq struct {
	Avatar string `json:"avatar" binding:"required,max=1024" comment:"头像相对路径"`
}

type ResetPasswordReq struct {
	ID       uint   `json:"id" binding:"required" comment:"用户ID"`
	Password string `json:"password" binding:"required,min=6" comment:"新密码"`
}
