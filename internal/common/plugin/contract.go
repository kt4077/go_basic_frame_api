// Package plugin 定义源码级插件的稳定契约、注册中心和生命周期。
// 插件只能使用宿主提供的路由组注册接口，不能绕过既有鉴权与审计中间件。
package plugin

import (
	"context"

	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
)

// ServiceType 插件当前运行的服务类型。
type ServiceType string

const (
	// ServiceAdmin 表示管理端 API 进程。
	ServiceAdmin ServiceType = "admin"
	// ServiceAPI 表示用户端 API 进程。
	ServiceAPI ServiceType = "api"
)

// CoreVersion 是当前插件契约兼容的核心版本。
const CoreVersion = "0.0.2"

// Dependency 插件依赖及其版本要求。
type Dependency struct {
	PluginID string `json:"plugin_id" comment:"依赖插件唯一标识"`
	Version  string `json:"version" comment:"依赖插件版本，第一阶段要求精确匹配"`
}

// Manifest 插件编译时清单。
type Manifest struct {
	PluginID     string       `json:"plugin_id" comment:"插件唯一标识"`
	Name         string       `json:"name" comment:"插件名称"`
	Version      string       `json:"version" comment:"插件版本"`
	Logo         string       `json:"logo" comment:"插件Logo相对路径"`
	Author       string       `json:"author" comment:"插件作者"`
	Homepage     string       `json:"homepage" comment:"插件主页地址"`
	Description  string       `json:"description" comment:"插件说明"`
	CoreVersion  string       `json:"core_version" comment:"兼容的核心版本说明"`
	Dependencies []Dependency `json:"dependencies" comment:"插件依赖"`
}

// Migration 描述插件要求存在的数据库迁移。
// SQL 仍由 sql 目录中的脚本显式执行，运行时只做版本和摘要校验。
type Migration struct {
	Version     string `json:"version" comment:"迁移版本"`
	Description string `json:"description" comment:"迁移说明"`
	Checksum    string `json:"checksum" comment:"迁移内容SHA256摘要"`
}

// AdminRouteGroups 管理端插件可用的受控路由组。
type AdminRouteGroups struct {
	Public     *gin.RouterGroup
	Auth       *gin.RouterGroup
	Permission *gin.RouterGroup
}

// APIRouteGroups 用户端插件可用的受控路由组。
type APIRouteGroups struct {
	Public *gin.RouterGroup
	Auth   *gin.RouterGroup
}

// Context 是宿主提供给插件的运行时上下文。
type Context struct {
	App *app.App
}

// Plugin 是源码级插件必须实现的后端契约。
type Plugin interface {
	Manifest() Manifest
	Migrations() []Migration
	RegisterAdminRoutes(ctx *Context, groups AdminRouteGroups) error
	RegisterAPIRoutes(ctx *Context, groups APIRouteGroups) error
	Start(ctx context.Context, service ServiceType) error
	Stop(ctx context.Context) error
}
