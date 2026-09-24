// Package plugins 维护编译进当前程序的源码级插件清单。
// 新插件必须在此显式注册，数据库启用状态不能让未编译代码直接运行。
package plugins

import commonplugin "server_api/internal/common/plugin"

// RegisterBuiltins 注册编译进程序的插件。当前版本只提供基础设施，
// 后续商城、新闻等插件在此加入，核心业务模块不需要迁移到本目录。
func RegisterBuiltins(registry *commonplugin.Registry) error {
	for _, item := range generatedBuiltins() {
		if err := registry.Register(item); err != nil {
			return err
		}
	}
	return nil
}
