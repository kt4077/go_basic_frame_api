package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/permission"
	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	"server_api/pkg/response"
)

// Permission 是管理端接口级权限中间件，按 请求方法:路由模板 与菜单权限匹配。
func Permission(application *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := auth.CtxClaims(c)
		if claims == nil {
			response.FailWithStatus(c, http.StatusUnauthorized, response.CodeErrAuth, "未登录")
			c.Abort()
			return
		}
		user, err := auth.GetUserByID(application.DB, claims.UserID)
		if err != nil || user.Status != enums.StatusEnabled {
			if err == nil {
				_ = auth.KickUser(application, user.ID)
			}
			response.FailWithStatus(c, http.StatusUnauthorized, response.CodeErrAuth, "账号不可用，请重新登录")
			c.Abort()
			return
		}
		if enums.SkipPermissionApis[c.Request.Method+":"+c.FullPath()] {
			c.Next()
			return
		}
		allowed, err := permission.HasPermission(c.Request.Context(), application, user, c.Request.Method, c.FullPath())
		if err != nil {
			log.Printf("权限校验异常 user_id=%d method=%s path=%s err=%v", user.ID, c.Request.Method, c.FullPath(), err)
			response.FailWithStatus(c, http.StatusServiceUnavailable, response.CodeErrServiceUnavailable, "权限服务暂时不可用，请稍后重试")
			c.Abort()
			return
		}
		if !allowed {
			response.FailWithStatus(c, http.StatusForbidden, response.CodeErrForbidden, "无权限执行该操作")
			c.Abort()
			return
		}
		c.Next()
	}
}
