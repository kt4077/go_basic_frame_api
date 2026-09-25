// Package router 接口端（用户端）路由，与管理端路由分文件管理。
// 除文件访问外，所有接口都必须携带平台来源与版本号请求头。
package router

import (
	"github.com/gin-gonic/gin"

	apicontroller "server_api/internal/api/controller"
	apilogic "server_api/internal/api/logic"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	commonmiddleware "server_api/internal/common/middleware"
	commonplugin "server_api/internal/common/plugin"
	commonupload "server_api/internal/common/upload"
)

// ApiRoutes 用户端路由。
func ApiRoutes(application *app.App, pluginRegistry *commonplugin.Registry) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), commonmiddleware.CORS())
	r.GET("/files/*filepath", func(c *gin.Context) { commonupload.ServeLocalFile(application, c) })

	smsLogic := &apilogic.SmsLogic{App: application}
	authC := &apicontroller.AuthController{Logic: &apilogic.AuthLogic{App: application, SMS: smsLogic}}
	smsC := &apicontroller.SmsController{Logic: smsLogic}
	uploadC := &apicontroller.UploadController{App: application}

	// 无需登录：短信验证码为通用能力，独立于认证接口
	sms := r.Group("/api/sms", commonmiddleware.Platform())
	{
		sms.POST("/code", smsC.SendCode)
	}

	// 无需登录：注册与登录能力
	pub := r.Group("/api", commonmiddleware.Platform())
	{
		pub.POST("/auth/register", authC.Register)
		pub.POST("/auth/login_password", authC.LoginPassword)
		pub.POST("/auth/login_sms", authC.LoginSms)
	}

	// 需登录：仅接受会员令牌
	auth := r.Group("/api/auth", commonmiddleware.Platform(), commonmiddleware.Auth(application, enums.ClientApi))
	{
		auth.POST("/logout", authC.Logout)
		auth.GET("/profile", authC.Profile)
		auth.POST("/profile/update", authC.UpdateProfile)
		auth.POST("/account/update", authC.UpdateAccount)
		auth.POST("/password/update", authC.UpdatePassword)
		auth.POST("/mobile/update", authC.UpdateMobile)
		auth.POST("/upload/file", uploadC.Upload)
	}

	if err := pluginRegistry.RegisterAPIRoutes(commonplugin.APIRouteGroups{Public: pub, Auth: auth}); err != nil {
		return nil, err
	}
	return r, nil
}
