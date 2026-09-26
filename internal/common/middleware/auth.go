// Package middleware 提供 admin 与 api 共用的 Gin 中间件。
package middleware

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/pkg/response"
)

// Auth JWT 鉴权中间件：校验 token，并检查 Redis 中 login_id 是否有效（主动过期）。
// allow 用于限定令牌的客户端类型（admin | api），防止管理端与用户端令牌互相调用。
func Auth(application *app.App, allow ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := auth.BearerToken(c)
		if tokenStr == "" {
			response.FailWithStatus(c, 401, response.CodeErrAuth, "未登录")
			c.Abort()
			return
		}
		claims, err := auth.ParseToken(application, tokenStr)
		if err != nil {
			response.FailWithStatus(c, 401, response.CodeErrAuth, err.Error())
			c.Abort()
			return
		}
		if !auth.ClientAllowed(claims, allow) {
			response.FailWithStatus(c, 401, response.CodeErrAuth, "登录已失效，请重新登录")
			c.Abort()
			return
		}
		c.Set(auth.CtxClaimsKey, claims)
		c.Next()
	}
}
