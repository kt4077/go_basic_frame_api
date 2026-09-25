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
	"server_api/internal/common/enums"
	commonmiddleware "server_api/internal/common/middleware"
	commonplugin "server_api/internal/common/plugin"
	commonupload "server_api/internal/common/upload"
)

// AdminRoutes 管理端路由。
func AdminRoutes(application *app.App, pluginRegistry *commonplugin.Registry) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), commonmiddleware.CORS())
	r.GET("/files/*filepath", func(c *gin.Context) { commonupload.ServeLocalFile(application, c) })

	authC := &controller.AuthController{Logic: &logic.AuthLogic{App: application}}
	menuC := &controller.MenuController{Logic: &logic.MenuLogic{App: application}}
	roleC := &controller.RoleController{Logic: &logic.RoleLogic{App: application}}
	userC := &controller.UserController{Logic: &logic.UserLogic{App: application}}
	memberC := &controller.MemberController{Logic: &logic.MemberLogic{App: application}}
	deptC := &controller.DeptController{Logic: &logic.DeptLogic{App: application}}
	storageC := &controller.StorageController{Logic: &logic.StorageLogic{App: application}}
	uploadC := &controller.UploadController{App: application}
	logC := &controller.OperationLogController{Logic: &logic.OperationLogLogic{App: application}}
	dashC := &controller.DashboardController{Logic: &logic.DashboardLogic{App: application}}
	smsC := &controller.SMSController{Logic: &logic.SMSLogic{App: application}}
	wechatC := &controller.WechatController{Logic: &logic.WechatLogic{App: application}}
	paymentC := &controller.PaymentController{Logic: &logic.PaymentLogic{App: application}}
	platformC := &controller.PlatformController{Logic: &logic.PlatformLogic{App: application}}
	pluginC := &controller.PluginController{Logic: &logic.PluginLogic{App: application, Registry: pluginRegistry}}

	// 无需登录
	pub := r.Group("/admin")
	{
		pub.POST("/login", authC.Login)
		pub.GET("/platform/public", platformC.AdminDetail)
	}

	// 需登录
	auth := r.Group("/admin", commonmiddleware.Auth(application, enums.ClientAdmin), adminmiddleware.OperationLog(application))
	{
		auth.POST("/logout", authC.Logout)
		auth.GET("/me", authC.Me)
		auth.POST("/profile/update", authC.UpdateProfile)
		auth.POST("/profile/avatar", authC.UpdateAvatar)
		auth.GET("/routers", authC.GetRouters)
		auth.GET("/permissions", authC.Permissions)
		auth.POST("/change_password", authC.ChangePassword)
	}

	// 需登录 + 接口级权限（api_path 与菜单/按钮绑定）
	perm := r.Group("/admin", commonmiddleware.Auth(application, enums.ClientAdmin), adminmiddleware.Permission(application), adminmiddleware.OperationLog(application))
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

		// 用户管理：系统用户
		perm.GET("/member/list", memberC.List)
		perm.POST("/member/set_status", memberC.SetStatus)

		// 文件上传（通用接口，免接口鉴权：见 enums.SkipPermissionApis）
		perm.POST("/upload/file", uploadC.Upload)

		// 日志维护
		perm.GET("/log/operation/list", logC.List)
		perm.POST("/log/operation/delete", logC.Delete)
		perm.POST("/log/operation/clear", logC.Clear)

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

		// 平台配置：管理端与用户端权限相互隔离
		perm.GET("/platform/admin/detail", platformC.AdminDetail)
		perm.POST("/platform/admin/save", platformC.SaveAdmin)
		perm.GET("/platform/user/detail", platformC.UserDetail)
		perm.POST("/platform/user/save", platformC.SaveUser)

		// 部门管理
		perm.GET("/dept/tree", deptC.Tree)
		perm.POST("/dept/add", deptC.Create)
		perm.POST("/dept/update", deptC.Update)
		perm.POST("/dept/delete", deptC.Delete)

		// 插件管理：状态修改后在服务重启时生效
		perm.GET("/plugin/list", pluginC.List)
		perm.GET("/plugin/detail", pluginC.Detail)
		perm.POST("/plugin/info", pluginC.UpdateInfo)
		perm.POST("/plugin/status", pluginC.UpdateStatus)
	}

	if err := pluginRegistry.RegisterAdminRoutes(commonplugin.AdminRouteGroups{
		Public: pub, Auth: auth, Permission: perm,
	}); err != nil {
		return nil, err
	}
	return r, nil
}
