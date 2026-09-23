package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/response"
)

const operationLogListPath = "/admin/log/operation/list"

// logBodyWriter 包装响应写入器，捕获接口响应内容
type logBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *logBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// OperationLog 操作日志中间件：记录管理端除登录外的所有接口调用，
// 包含请求参数（敏感字段脱敏）与接口响应内容（超长截断）。
func OperationLog(application *app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 登录及操作日志查询接口不记录，避免查看日志时产生新的操作日志。
		if shouldSkipOperationLog(c.Request.Method, c.FullPath()) {
			c.Next()
			return
		}
		start := time.Now()

		// 请求参数：JSON 体预读脱敏后回填，保证后续 handler 正常绑定
		var reqParams string
		if isJSONBody(c) && c.Request.Body != nil {
			raw, _ := io.ReadAll(c.Request.Body)
			_ = c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(raw))
			reqParams = maskParams(raw)
		}

		// 捕获响应内容
		bw := &logBodyWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = bw

		c.Next()

		// 非 JSON 请求：记录查询串 / 表单字段 / 上传文件名
		if reqParams == "" {
			reqParams = readRawParams(c)
		}

		entry := &model.SysOperationLog{
			Method:         c.Request.Method,
			Path:           c.FullPath(),
			RequestParams:  reqParams,
			ResponseParams: bw.body.String(),
			IP:             c.ClientIP(),
			UserAgent:      c.GetHeader("User-Agent"),
			Code:           readBizCode(c),
			CostMs:         time.Since(start).Milliseconds(),
			Client:         enums.ClientAdmin,
		}
		if claims := auth.CtxClaims(c); claims != nil {
			userID := claims.UserID
			entry.UserID = &userID
			entry.Username = claims.Username
		}
		if err := application.DB.Create(entry).Error; err != nil {
			log.Println("操作日志写入失败:", err)
		}
	}
}

// shouldSkipOperationLog 判断当前接口是否应跳过操作日志记录。
// 使用请求方法与完整路由共同匹配，避免同一路径下其他写操作被意外忽略。
func shouldSkipOperationLog(method, path string) bool {
	if method == "POST" && path == "/admin/login" {
		return true
	}
	return method == "GET" && path == operationLogListPath
}

func isJSONBody(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Content-Type"), "application/json")
}

// readRawParams 非 JSON 请求的参数：查询串 / 表单字段 / 上传文件名
func readRawParams(c *gin.Context) string {
	var parts []string
	if q := c.Request.URL.RawQuery; q != "" {
		parts = append(parts, q)
	}
	if c.Request.PostForm != nil {
		for k, vs := range c.Request.PostForm {
			parts = append(parts, k+"="+strings.Join(vs, ","))
		}
	}
	if c.Request.MultipartForm != nil {
		for k, fs := range c.Request.MultipartForm.File {
			names := make([]string, 0, len(fs))
			for _, f := range fs {
				names = append(names, f.Filename)
			}
			parts = append(parts, k+"="+strings.Join(names, ","))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "&")
}

// maskParams 请求参数脱敏：password/token/secret 类字段替换为掩码（不截断，全量保存）
func maskParams(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	body := string(raw)
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err == nil {
		for key := range m {
			lk := strings.ToLower(key)
			for _, sensitive := range enums.SensitiveParamKeys {
				if strings.Contains(lk, sensitive) {
					m[key] = "******"
					break
				}
			}
		}
		if b, err := json.Marshal(m); err == nil {
			body = string(b)
		}
	}
	return body
}

// readBizCode 从 context 读取业务响应码（由 response.OK/Fail 写入）
func readBizCode(c *gin.Context) int {
	if v, exists := c.Get(response.CtxBizCodeKey); exists {
		if code, ok := v.(int); ok {
			return code
		}
	}
	return 0
}
