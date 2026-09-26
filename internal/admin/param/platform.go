package param

// AdminPlatformSaveReq 管理端平台配置保存参数。
type AdminPlatformSaveReq struct {
	Logo       string `json:"logo" binding:"omitempty,max=500" validate:"管理端Logo相对路径" comment:"管理端Logo相对路径"`
	SystemName string `json:"system_name" binding:"required,max=100" validate:"系统名称" comment:"系统名称"`
}

// UserPlatformSaveReq 用户端平台配置保存参数。
type UserPlatformSaveReq struct {
	DefaultNickname string `json:"default_nickname" binding:"required,max=100" validate:"默认昵称" comment:"默认昵称"`
	DefaultAvatar   string `json:"default_avatar" binding:"omitempty,max=500" validate:"默认头像相对路径" comment:"默认头像相对路径"`
}
