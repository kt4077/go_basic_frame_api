package model

// SysMenu 菜单/按钮，parent_id 支持无限级。
// ApiPath 记录该菜单/按钮对应的后端接口地址，后端按此做接口级鉴权。
// 类型与状态枚举见 internal/common/enums/menu.go。表：sys_menu
type SysMenu struct {
	Base
	Name     string `gorm:"size:32;not null;comment:菜单名称" json:"name"`
	Type     int    `gorm:"default:2;index;comment:菜单类型" json:"type"`
	ParentID uint   `gorm:"default:0;index;comment:上级菜单ID，0表示顶级" json:"parent_id"`
	Path     string `gorm:"size:128;default:'';comment:前端路由地址" json:"path"`
	ApiPath  string `gorm:"size:255;default:'';comment:后端接口权限地址" json:"api_path"`
	Icon     string `gorm:"size:64;default:'';comment:菜单图标" json:"icon"`
	Sort     int    `gorm:"default:0;comment:排序" json:"sort"`
	Status   int    `gorm:"default:1;comment:显示状态" json:"status"`
	Remark   string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysMenu) TableName() string { return "sys_menu" }

// TreeNode 接口实现：可用于 pkg/tree 的树构建与后代收集。
func (m SysMenu) GetID() uint       { return m.ID }
func (m SysMenu) GetParentID() uint { return m.ParentID }
