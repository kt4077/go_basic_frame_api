// Package middleware 提供 admin 与 api 共用的 Gin 中间件。
package middleware

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/pkg/response"
)

// Auth JWT 鉴权中间件：校验 token，并检查 Redis 中 login_id 是否有效（主动过期）。
func Auth(application *app.App) gin.HandlerFunc {
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
		c.Set(auth.CtxClaimsKey, claims)
		c.Next()
	}
}
