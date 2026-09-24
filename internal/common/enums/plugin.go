package enums

// 插件状态枚举，从 1 开始定义。
const (
	PluginStatusEnabled  = 1 // 启用
	PluginStatusDisabled = 2 // 停用
)

// 插件迁移执行状态枚举，从 1 开始定义。
const (
	PluginMigrationSuccess = 1 // 执行成功
	PluginMigrationFailed  = 2 // 执行失败
)
