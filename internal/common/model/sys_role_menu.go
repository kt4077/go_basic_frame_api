package model

// SysRoleMenu 角色-菜单(含按钮)关联。表：sys_role_menu
type SysRoleMenu struct {
	RoleID uint `gorm:"primaryKey;comment:角色ID" json:"role_id"`
	MenuID uint `gorm:"primaryKey;comment:菜单ID" json:"menu_id"`
}

func (SysRoleMenu) TableName() string { return "sys_role_menu" }
