package model

// SysPluginMenu 保存插件菜单业务键与系统菜单自增ID的映射。
// 插件安装和升级只使用业务键描述层级，不允许指定 sys_menu.id。表：sys_plugin_menu
type SysPluginMenu struct {
	ID           uint   `gorm:"primaryKey;comment:主键ID" json:"id"`
	PluginID     string `gorm:"size:64;not null;uniqueIndex:uk_plugin_menu_key;comment:插件唯一标识" json:"plugin_id"`
	MenuKey      string `gorm:"size:64;not null;uniqueIndex:uk_plugin_menu_key;comment:插件内菜单业务键" json:"menu_key"`
	MenuID       uint   `gorm:"not null;uniqueIndex:uk_plugin_menu_id;comment:系统菜单ID" json:"menu_id"`
	ParentKey    string `gorm:"size:64;not null;default:'';comment:插件内父级菜单业务键" json:"parent_key"`
	ParentSource int    `gorm:"not null;default:1;comment:父级来源，1插件默认，2管理员自定义" json:"parent_source"`
}

// TableName 返回插件菜单映射表名。
func (SysPluginMenu) TableName() string { return "sys_plugin_menu" }
