package enums

// SensitiveParamKeys 操作日志参数脱敏：请求参数中命中的 key 值替换为掩码
var SensitiveParamKeys = []string{
	"password", "token", "secret",
	"private_key", "public_key", "api_v3_key", "aes_key",
}
