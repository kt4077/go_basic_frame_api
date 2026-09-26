// Package news 提供新闻分类、文章发布和用户端新闻查询能力。
package news

import (
	"context"

	commonplugin "server_api/internal/common/plugin"
	adminController "server_api/internal/plugins/news/admin/controller"
	adminLogic "server_api/internal/plugins/news/admin/logic"
	apiController "server_api/internal/plugins/news/api/controller"
	apiLogic "server_api/internal/plugins/news/api/logic"
)

const (
	migrationV100Checksum = "e08ece96f018238258917abc862356d6c30e1c1679e28f5ddd89c19d8dbf5a13"
	migrationV110Checksum = "092fde34d3ccfbe517eba4ab0abca05e588c7138c75e006b06bd8503de6a4c27"
)

// Plugin 新闻资讯插件。
type Plugin struct{}

// New 创建新闻资讯插件。
func New() *Plugin { return &Plugin{} }

// Manifest 返回插件元信息。
func (p *Plugin) Manifest() commonplugin.Manifest {
	return commonplugin.Manifest{PluginID: "news", Name: "新闻资讯", Version: "1.1.4", Logo: "", Author: "kt4077", Homepage: "https://github.com/kt4077/go_basic_frame_plugin_news", Description: "提供新闻分类、文章草稿、发布下线、封面管理以及用户端新闻查询能力。", CoreVersion: commonplugin.CoreVersion, Dependencies: []commonplugin.Dependency{}}
}

// Migrations 返回插件迁移声明。
func (p *Plugin) Migrations() []commonplugin.Migration {
	return []commonplugin.Migration{
		{Version: "1.0.0", Description: "创建新闻分类与文章表", Checksum: migrationV100Checksum},
		{Version: "1.1.0", Description: "增加文章启用状态与虚拟浏览数量", Checksum: migrationV110Checksum},
	}
}

// RegisterAdminRoutes 注册受鉴权和权限校验保护的管理端路由。
func (p *Plugin) RegisterAdminRoutes(ctx *commonplugin.Context, groups commonplugin.AdminRouteGroups) error {
	category := &adminController.CategoryController{Logic: &adminLogic.CategoryLogic{App: ctx.App}}
	article := &adminController.ArticleController{Logic: &adminLogic.ArticleLogic{App: ctx.App}}
	group := groups.Permission.Group("/plugin/news")
	group.GET("/category/list", category.List)
	group.POST("/category/save", category.Save)
	group.POST("/category/delete", category.Delete)
	group.GET("/article/list", article.List)
	group.GET("/article/detail", article.Detail)
	group.POST("/article/save", article.Save)
	group.POST("/article/status", article.UpdateStatus)
	group.POST("/article/delete", article.Delete)
	return nil
}

// RegisterAPIRoutes 注册用户端公开新闻路由。
func (p *Plugin) RegisterAPIRoutes(ctx *commonplugin.Context, groups commonplugin.APIRouteGroups) error {
	controller := &apiController.NewsController{Logic: &apiLogic.NewsLogic{App: ctx.App}}
	group := groups.Public.Group("/plugin/news")
	group.GET("/categories", controller.Categories)
	group.GET("/articles", controller.Articles)
	group.GET("/articles/:slug", controller.Article)
	return nil
}

// Start 启动插件，本版本没有后台任务。
func (p *Plugin) Start(ctx context.Context, service commonplugin.ServiceType) error { return nil }

// Stop 停止插件，本版本没有后台任务。
func (p *Plugin) Stop(ctx context.Context) error { return nil }
