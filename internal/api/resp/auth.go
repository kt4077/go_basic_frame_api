// Package resp 用户端响应参数定义，按业务功能组织文件。
package resp

import (
	"time"

	"server_api/internal/common/model"
	"server_api/pkg/mask"
)

type LoginRes struct {
	Token string `json:"token" comment:"访问令牌"`
}

type ProfileRes struct {
	User UserItem `json:"user" comment:"用户信息"`
	Dept DeptItem `json:"dept" comment:"部门信息"`
}

type UserItem struct {
	ID        uint      `json:"id" comment:"主键ID"`
	CreatedAt time.Time `json:"created_at" comment:"创建时间"`
	UpdatedAt time.Time `json:"updated_at" comment:"更新时间"`
	Username  string    `json:"username" comment:"登录账号"`
	Nickname  string    `json:"nickname" comment:"用户昵称"`
	Avatar    string    `json:"avatar" comment:"头像地址"`
	Mobile    string    `json:"mobile" comment:"手机号"`
	Email     string    `json:"email" comment:"邮箱"`
	DeptID    uint      `json:"dept_id" comment:"部门ID"`
	Status    int       `json:"status" comment:"状态"`
	IsSuper   int       `json:"is_super" comment:"是否超级管理员"`
}

type DeptItem struct {
	ID       uint   `json:"id" comment:"主键ID"`
	Name     string `json:"name" comment:"部门名称"`
	ParentID uint   `json:"parent_id" comment:"上级部门ID"`
	Sort     int    `json:"sort" comment:"排序"`
	Leader   string `json:"leader" comment:"负责人"`
	Remark   string `json:"remark" comment:"备注"`
}

func NewUserItem(v model.SysUser) UserItem {
	return UserItem{ID: v.ID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt, Username: v.Username, Nickname: v.Nickname, Avatar: v.Avatar, Mobile: mask.Mobile(v.Mobile), Email: v.Email, DeptID: v.DeptID, Status: v.Status, IsSuper: v.IsSuper}
}

func NewDeptItem(v model.SysDept) DeptItem {
	return DeptItem{ID: v.ID, Name: v.Name, ParentID: v.ParentID, Sort: v.Sort, Leader: v.Leader, Remark: v.Remark}
}
