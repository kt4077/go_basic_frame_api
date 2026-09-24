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

// 插件安装操作类型枚举，从 1 开始定义。
const (
	PluginInstallActionInstall = 1 // 安装
	PluginInstallActionUpgrade = 2 // 升级
)

// 插件安装执行状态枚举，从 1 开始定义。
const (
	PluginInstallStatusSuccess = 1 // 执行成功
	PluginInstallStatusFailed  = 2 // 执行失败
)
