package enums

// 登录流水状态
const (
	LoginStatusOnline = 1 // 在线
	LoginStatusLogout = 2 // 已退出
	LoginStatusKicked = 3 // 被踢下线/禁用
)

// 客户端类型
const (
	ClientAdmin = "admin" // 管理端
	ClientApi   = "api"   // 用户端
)

// 超级管理员标志
const (
	IsSuperYes = 1
	IsSuperNo  = 0
)
