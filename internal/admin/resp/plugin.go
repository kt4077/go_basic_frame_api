package resp

import "time"

// PluginItem 插件列表项。
type PluginItem struct {
	BaseItem    `comment:"基础字段"`
	PluginID    string `json:"plugin_id" comment:"插件唯一标识"`
	Name        string `json:"name" comment:"插件名称"`
	Version     string `json:"version" comment:"数据库安装版本"`
	CodeVersion string `json:"code_version" gorm:"-" comment:"当前程序编译版本"`
	Logo        string `json:"logo" comment:"插件Logo相对路径"`
	LogoURL     string `json:"logo_url" gorm:"-" comment:"插件Logo完整访问地址"`
	Author      string `json:"author" comment:"插件作者"`
	Homepage    string `json:"homepage" comment:"插件主页地址"`
	Description string `json:"description" comment:"插件描述"`
	Status      int    `json:"status" comment:"插件状态，1启用，2停用"`
	Compiled    bool   `json:"compiled" gorm:"-" comment:"是否编译进当前程序"`
}

// PluginMigrationItem 插件迁移记录。
type PluginMigrationItem struct {
	ID           uint      `json:"id" comment:"主键ID"`
	Version      string    `json:"version" comment:"迁移版本"`
	Checksum     string    `json:"checksum" comment:"迁移内容SHA256摘要"`
	Status       int       `json:"status" comment:"迁移状态，1成功，2失败"`
	ExecutionMs  int64     `json:"execution_ms" comment:"执行耗时毫秒"`
	ErrorMessage string    `json:"error_message" comment:"失败原因"`
	ExecutedAt   time.Time `json:"executed_at" comment:"执行时间"`
}

// PluginDetail 插件详情响应。
type PluginDetail struct {
	PluginItem `comment:"插件基础信息"`
	Manifest   string                `json:"manifest" comment:"安装时插件清单JSON快照"`
	Migrations []PluginMigrationItem `json:"migrations" comment:"数据库迁移记录"`
}

// PluginListRes 插件分页列表响应。
type PluginListRes struct {
	List  []PluginItem `json:"list" comment:"插件列表"`
	Total int64        `json:"total" comment:"数据总数"`
}
