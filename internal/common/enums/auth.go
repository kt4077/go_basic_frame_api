package enums

// SkipPermissionApis 免接口权限校验白名单（key = "METHOD:/路由pattern"）。
// 通用能力接口（如文件上传）不与菜单/按钮绑定，加入本 map 后，
// Permission 中间件直接放行，避免出现权限不足。
// 说明：用户端(/api)未启用接口鉴权中间件，因此只需登记管理端的接口。
var SkipPermissionApis = map[string]bool{
	"POST:/admin/upload/file":   true, // multipart 表单上传
	"POST:/admin/upload/stream": true, // 二进制文件流上传
}
