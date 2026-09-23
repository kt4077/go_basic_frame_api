package model

// SysRole 角色，parent_id 支持多级；上级角色自动拥有子级、孙级角色的权限。表：sys_role
type SysRole struct {
	Base
	Name     string `gorm:"size:32;not null;comment:角色名称" json:"name"`
	Code     string `gorm:"size:32;not null;uniqueIndex;comment:角色编码" json:"code"`
	ParentID uint   `gorm:"default:0;index;comment:上级角色ID，0表示顶级" json:"parent_id"`
	Sort     int    `gorm:"default:0;comment:排序" json:"sort"`
	Status   int    `gorm:"default:1;comment:状态" json:"status"`
	Remark   string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysRole) TableName() string { return "sys_role" }

// TreeNode 接口实现：可用于 pkg/tree 的树构建与后代收集。
func (r SysRole) GetID() uint       { return r.ID }
func (r SysRole) GetParentID() uint { return r.ParentID }
