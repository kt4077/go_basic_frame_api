package resp

// AdminPlatformConfigRes 管理端平台配置响应。
type AdminPlatformConfigRes struct {
	Logo       string `json:"logo" comment:"管理端Logo访问地址"`
	LogoPath   string `json:"logo_path" comment:"管理端Logo相对路径"`
	SystemName string `json:"system_name" comment:"系统名称"`
}

// UserPlatformConfigRes 用户端平台配置响应。
type UserPlatformConfigRes struct {
	DefaultNickname   string `json:"default_nickname" comment:"默认昵称"`
	DefaultAvatar     string `json:"default_avatar" comment:"默认头像访问地址"`
	DefaultAvatarPath string `json:"default_avatar_path" comment:"默认头像相对路径"`
}
