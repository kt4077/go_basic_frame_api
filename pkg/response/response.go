// Package response 提供 Gin 接口统一响应格式。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应码
const (
	CodeOK                    = 0
	CodeErrParams             = 400
	CodeErrAuth               = 401 // 未登录或登录已失效
	CodeErrForbidden          = 403 // 无权限
	CodeErrBusiness           = 500
	CodeErrServiceUnavailable = 503 // 依赖服务暂时不可用
)

// CtxBizCodeKey 业务响应码在 gin context 中的 key（操作日志中间件读取）
const CtxBizCodeKey = "biz_code"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func OK(c *gin.Context, data interface{}) {
	c.Set(CtxBizCodeKey, CodeOK)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, Response{Code: CodeOK, Msg: "success", Data: data})
}

func Fail(c *gin.Context, code int, msg string) {
	c.Set(CtxBizCodeKey, code)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}

func FailWithStatus(c *gin.Context, httpStatus, code int, msg string) {
	c.Set(CtxBizCodeKey, code)
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(httpStatus, Response{Code: code, Msg: msg})
}
