package param

type UserListReq struct {
	Keyword  string `form:"keyword" validate:"搜索关键字" comment:"搜索关键字"`
	Status   int    `form:"status" validate:"状态" comment:"状态"`
	DeptID   uint   `form:"dept_id" validate:"部门ID" comment:"部门ID"`
	Page     int    `form:"page" validate:"页码" comment:"页码"`
	PageSize int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}

type UserSaveReq struct {
	ID       uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Username string `json:"username" validate:"登录账号" comment:"登录账号"`
	Password string `json:"password" validate:"登录密码" comment:"登录密码"`
	Nickname string `json:"nickname" validate:"用户昵称" comment:"用户昵称"`
	Avatar   string `json:"avatar" binding:"omitempty,max=1024" validate:"头像路径" comment:"头像路径"`
	Mobile   string `json:"mobile" validate:"手机号" comment:"手机号"`
	Email    string `json:"email" validate:"邮箱" comment:"邮箱"`
	DeptID   uint   `json:"dept_id" validate:"部门ID" comment:"部门ID"`
	Status   int    `json:"status" validate:"状态" comment:"状态"`
	IsSuper  int    `json:"is_super" validate:"是否超级管理员" comment:"是否超级管理员"`
	RoleIDs  []uint `json:"role_ids" validate:"角色ID列表" comment:"角色ID列表"`
}

// ProfileUpdateReq 当前管理员可修改的个人资料。
type ProfileUpdateReq struct {
	Nickname string `json:"nickname" binding:"max=32" validate:"用户昵称" comment:"用户昵称"`
	Avatar   string `json:"avatar" binding:"omitempty,max=1024" validate:"头像路径" comment:"头像路径"`
	Mobile   string `json:"mobile" binding:"omitempty,max=16" validate:"手机号" comment:"手机号"`
	Email    string `json:"email" binding:"omitempty,email,max=64" validate:"邮箱" comment:"邮箱"`
}

// AvatarUpdateReq 当前管理员头像更新参数。
type AvatarUpdateReq struct {
	Avatar string `json:"avatar" binding:"required,max=1024" validate:"头像相对路径" comment:"头像相对路径"`
}

type ResetPasswordReq struct {
	ID       uint   `json:"id" binding:"required" validate:"用户ID" comment:"用户ID"`
	Password string `json:"password" binding:"required,min=6" validate:"新密码" comment:"新密码"`
}
