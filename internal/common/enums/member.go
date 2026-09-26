package enums

import "slices"

// 会员用户性别
const (
	GenderMale    = 1 // 男
	GenderFemale  = 2 // 女
	GenderUnknown = 3 // 未知
)

// ValidGenders 会员用户性别的合法取值，新增性别时在此登记。
var ValidGenders = []int{
	GenderMale,
	GenderFemale,
	GenderUnknown,
}

// IsValidGender 判断性别是否为已定义的合法枚举值，以枚举集合判断不使用数值范围。
func IsValidGender(gender int) bool {
	return slices.Contains(ValidGenders, gender)
}

// 会员用户注册来源
const (
	RegisterSourceWechatMini = 1 // 微信小程序
	RegisterSourceWechatOA   = 2 // 微信公众号
	RegisterSourceIOS        = 3 // iOS
	RegisterSourceAndroid    = 4 // Android
)

// ValidRegisterSources 会员用户注册来源的合法取值，新增来源时在此登记。
var ValidRegisterSources = []int{
	RegisterSourceWechatMini,
	RegisterSourceWechatOA,
	RegisterSourceIOS,
	RegisterSourceAndroid,
}

// IsValidRegisterSource 判断注册来源是否为已定义的合法枚举值。
// 以枚举集合判断，不使用数值范围，避免新增来源后范围判断失效。
func IsValidRegisterSource(source int) bool {
	return slices.Contains(ValidRegisterSources, source)
}
