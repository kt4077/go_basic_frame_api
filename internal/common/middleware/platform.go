package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"server_api/internal/common/enums"
	"server_api/pkg/response"
)

// 平台来源与版本号请求头：用户端每个接口都需要携带，来源取值与会员注册来源一致。
const (
	HeaderPlatformSource = "X-Platform-Source"
	HeaderAppVersion     = "X-App-Version"

	CtxPlatformSourceKey = "platform_source"
	CtxAppVersionKey     = "app_version"
)

// Platform 解析平台来源与版本号请求头并写入上下文。
// 来源必须是 enums 中已注册的会员注册来源枚举值；
// 版本号仅透传不做一致性校验，供后续灰度与兼容判断使用。
func Platform() gin.HandlerFunc {
	return func(c *gin.Context) {
		source, err := strconv.Atoi(c.GetHeader(HeaderPlatformSource))
		if err != nil || !enums.IsValidRegisterSource(source) {
			response.Fail(c, response.CodeErrParams, "平台来源缺失或不合法")
			c.Abort()
			return
		}
		c.Set(CtxPlatformSourceKey, source)
		c.Set(CtxAppVersionKey, c.GetHeader(HeaderAppVersion))
		c.Next()
	}
}

// CtxPlatformSource 从上下文读取平台来源。
func CtxPlatformSource(c *gin.Context) int {
	v, ok := c.Get(CtxPlatformSourceKey)
	if !ok {
		return 0
	}
	source, _ := v.(int)
	return source
}

// CtxAppVersion 从上下文读取客户端版本号。
func CtxAppVersion(c *gin.Context) string {
	v, ok := c.Get(CtxAppVersionKey)
	if !ok {
		return ""
	}
	version, _ := v.(string)
	return version
}
