package model

// SysUserRole 用户-角色关联。表：sys_user_role
type SysUserRole struct {
	UserID uint `gorm:"primaryKey;comment:用户ID" json:"user_id"`
	RoleID uint `gorm:"primaryKey;comment:角色ID" json:"role_id"`
}

func (SysUserRole) TableName() string { return "sys_user_role" }
