package model

// SysDept 公司组织架构，parent_id 支持多级。表：sys_dept
type SysDept struct {
	Base
	Name     string `gorm:"size:64;not null;comment:部门名称" json:"name"`
	ParentID uint   `gorm:"default:0;index;comment:上级部门ID，0表示顶级" json:"parent_id"`
	Sort     int    `gorm:"default:0;comment:排序" json:"sort"`
	Leader   string `gorm:"size:32;default:'';comment:负责人" json:"leader"`
	Remark   string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysDept) TableName() string { return "sys_dept" }

// TreeNode 接口实现：可用于 pkg/tree 的树构建与后代收集。
func (d SysDept) GetID() uint       { return d.ID }
func (d SysDept) GetParentID() uint { return d.ParentID }
