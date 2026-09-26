// Package mask 展示字段脱敏工具。
package mask

import "strings"

// Mobile 手机号脱敏：保留前 3 位与后 4 位，如 138****8001。
// 空串返回空串；非 11 位原样返回，避免误伤历史或异常数据。
func Mobile(mobile string) string {
	mobile = strings.TrimSpace(mobile)
	if mobile == "" {
		return ""
	}
	if len(mobile) != 11 {
		return mobile
	}
	return mobile[:3] + "****" + mobile[7:]
}

// MobilePtr 可空手机号脱敏：nil 或空值返回空串。
func MobilePtr(mobile *string) string {
	if mobile == nil {
		return ""
	}
	return Mobile(*mobile)
}
