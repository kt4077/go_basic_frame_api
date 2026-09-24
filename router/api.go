// Package router 接口端（用户端）路由，与管理端路由分文件管理。
package router

import (
	"github.com/gin-gonic/gin"

	apicontroller "server_api/internal/api/controller"
	apilogic "server_api/internal/api/logic"
	"server_api/internal/common/app"
	commonmiddleware "server_api/internal/common/middleware"
	commonplugin "server_api/internal/common/plugin"
	commonupload "server_api/internal/common/upload"
)

// ApiRoutes 用户端路由。
func ApiRoutes(application *app.App, pluginRegistry *commonplugin.Registry) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), commonmiddleware.CORS())
	r.GET("/files/*filepath", func(c *gin.Context) { commonupload.ServeLocalFile(application, c) })

	authC := &apicontroller.AuthController{Logic: &apilogic.AuthLogic{App: application}}
	uploadC := &apicontroller.UploadController{App: application}

	// 无需登录
	pub := r.Group("/api")
	{
		pub.POST("/login", authC.Login)
	}

	// 需登录
	auth := r.Group("/api", commonmiddleware.Auth(application))
	{
		auth.POST("/logout", authC.Logout)
		auth.POST("/upload/file", uploadC.Upload)
		auth.GET("/profile", authC.Profile)
		auth.POST("/change_password", authC.ChangePassword)
	}

	if err := pluginRegistry.RegisterAPIRoutes(commonplugin.APIRouteGroups{Public: pub, Auth: auth}); err != nil {
		return nil, err
	}
	return r, nil
}
