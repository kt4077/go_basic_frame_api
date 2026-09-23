// Package router 路由注册。管理端与接口端路由分文件管理：
// 管理端路由在本文件，接口端路由见 api.go。
// 路由采用统一资源 + 动作（action）风格：/admin/{模块}/{动作}。
package router

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/controller"
	"server_api/internal/admin/logic"
	adminmiddleware "server_api/internal/admin/middleware"
	"server_api/internal/common/app"
	commonmiddleware "server_api/internal/common/middleware"
	commonupload "server_api/internal/common/upload"
)

// AdminRoutes 管理端路由。
func AdminRoutes(application *app.App) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), commonmiddleware.CORS())
	r.GET("/files/*filepath", func(c *gin.Context) { commonupload.ServeLocalFile(application, c) })

	authC := &controller.AuthController{Logic: &logic.AuthLogic{App: application}}
	menuC := &controller.MenuController{Logic: &logic.MenuLogic{App: application}}
	roleC := &controller.RoleController{Logic: &logic.RoleLogic{App: application}}
	userC := &controller.UserController{Logic: &logic.UserLogic{App: application}}
	deptC := &controller.DeptController{Logic: &logic.DeptLogic{App: application}}
	storageC := &controller.StorageController{Logic: &logic.StorageLogic{App: application}}
	uploadC := &controller.UploadController{App: application}
	logC := &controller.OperationLogController{Logic: &logic.OperationLogLogic{App: application}}
	dashC := &controller.DashboardController{Logic: &logic.DashboardLogic{App: application}}
	smsC := &controller.SMSController{Logic: &logic.SMSLogic{App: application}}
	wechatC := &controller.WechatController{Logic: &logic.WechatLogic{App: application}}
	paymentC := &controller.PaymentController{Logic: &logic.PaymentLogic{App: application}}

	// 无需登录
	pub := r.Group("/admin")
	{
		pub.POST("/login", authC.Login)
	}

	// 需登录
	auth := r.Group("/admin", commonmiddleware.Auth(application), adminmiddleware.OperationLog(application))
	{
		auth.POST("/logout", authC.Logout)
		auth.GET("/me", authC.Me)
		auth.POST("/profile/update", authC.UpdateProfile)
		auth.GET("/routers", authC.GetRouters)
		auth.GET("/permissions", authC.Permissions)
		auth.POST("/change_password", authC.ChangePassword)
	}

	// 需登录 + 接口级权限（api_path 与菜单/按钮绑定）
	perm := r.Group("/admin", commonmiddleware.Auth(application), adminmiddleware.Permission(application), adminmiddleware.OperationLog(application))
	{
		// 系统总览
		perm.GET("/dashboard/overview", dashC.Overview)

		// 菜单管理
		perm.GET("/menu/list", menuC.List)
		perm.GET("/menu/tree", menuC.Tree)
		perm.POST("/menu/add", menuC.Create)
		perm.POST("/menu/update", menuC.Update)
		perm.POST("/menu/delete", menuC.Delete)

		// 角色管理
		perm.GET("/role/list", roleC.List)
		perm.GET("/role/tree", roleC.Tree)
		perm.POST("/role/add", roleC.Create)
		perm.POST("/role/update", roleC.Update)
		perm.POST("/role/delete", roleC.Delete)
		perm.GET("/role/menus", roleC.MenuIDs) // ?id= 角色已绑定的菜单ID
		perm.POST("/role/assign_menus", roleC.AssignMenus)
		perm.GET("/role/users", roleC.UserIDs) // ?id= 角色下的用户ID

		// 人员管理
		perm.GET("/user/list", userC.List)
		perm.POST("/user/add", userC.Create)
		perm.POST("/user/update", userC.Update)
		perm.POST("/user/delete", userC.Delete)
		perm.POST("/user/reset_password", userC.ResetPassword)
		perm.POST("/user/kick", userC.Kick)

		// 文件上传（通用接口，免接口鉴权：见 enums.SkipPermissionApis）
		perm.POST("/upload/file", uploadC.Upload)

		// 日志维护
		perm.GET("/log/operation/list", logC.List)

		// 存储渠道配置
		perm.GET("/storage/list", storageC.List)
		perm.POST("/storage/add", storageC.Create)
		perm.POST("/storage/update", storageC.Update)
		perm.POST("/storage/set_default", storageC.SetDefault)
		perm.POST("/storage/delete", storageC.Delete)

		// 短信配置：开发信息、签名、模板、发送记录
		perm.GET("/sms/config/list", smsC.ConfigList)
		perm.POST("/sms/config/save", smsC.SaveConfig)
		perm.POST("/sms/config/delete", smsC.DeleteConfig)
		perm.GET("/sms/signature/list", smsC.SignatureList)
		perm.POST("/sms/signature/save", smsC.SaveSignature)
		perm.POST("/sms/signature/delete", smsC.DeleteSignature)
		perm.GET("/sms/template/list", smsC.TemplateList)
		perm.POST("/sms/template/save", smsC.SaveTemplate)
		perm.POST("/sms/template/delete", smsC.DeleteTemplate)
		perm.GET("/sms/log/list", smsC.LogList)

		// 微信与支付渠道配置
		perm.GET("/wechat/config/list", wechatC.List)
		perm.POST("/wechat/config/save", wechatC.Save)
		perm.POST("/wechat/config/delete", wechatC.Delete)
		perm.GET("/payment/config/list", paymentC.List)
		perm.POST("/payment/config/save", paymentC.Save)
		perm.POST("/payment/config/delete", paymentC.Delete)

		// 部门管理
		perm.GET("/dept/tree", deptC.Tree)
		perm.POST("/dept/add", deptC.Create)
		perm.POST("/dept/update", deptC.Update)
		perm.POST("/dept/delete", deptC.Delete)
	}
	return r
}
