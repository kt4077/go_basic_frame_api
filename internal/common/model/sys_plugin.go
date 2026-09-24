package model

// SysPlugin 已安装插件及其当前状态。Manifest 保存安装时的插件清单快照，
// 运行时仍以编译进程序的插件清单为准。表：sys_plugin
type SysPlugin struct {
	Base
	PluginID    string `gorm:"size:64;not null;uniqueIndex:uk_sys_plugin_plugin_id;comment:插件唯一标识" json:"plugin_id"`
	Name        string `gorm:"size:100;not null;comment:插件名称" json:"name"`
	Version     string `gorm:"size:32;not null;comment:已安装插件版本" json:"version"`
	Logo        string `gorm:"size:500;not null;default:'';comment:插件Logo相对路径" json:"logo"`
	Author      string `gorm:"size:100;not null;default:'';comment:插件作者" json:"author"`
	Homepage    string `gorm:"size:500;not null;default:'';comment:插件主页地址" json:"homepage"`
	Description string `gorm:"size:2000;not null;default:'';comment:插件描述" json:"description"`
	Status      int    `gorm:"not null;default:2;index;comment:状态，1启用，2停用" json:"status"`
	Manifest    string `gorm:"type:mediumtext;comment:安装时插件清单JSON快照" json:"manifest"`
}

// TableName 返回插件信息表名。
func (SysPlugin) TableName() string { return "sys_plugin" }
