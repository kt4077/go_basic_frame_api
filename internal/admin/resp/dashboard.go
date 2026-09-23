package resp

type TrendItem struct {
	Name  string `json:"name" comment:"统计日期"`
	Value int64  `json:"value" comment:"统计数量"`
}

type DeptUserItem struct {
	DeptID uint   `json:"dept_id" comment:"部门ID"`
	Name   string `json:"name" comment:"部门名称"`
	Total  int64  `json:"total" comment:"用户数量"`
}

type OverviewRes struct {
	UserCount   int64             `json:"user_count" comment:"用户数量"`
	RoleCount   int64             `json:"role_count" comment:"角色数量"`
	MenuCount   int64             `json:"menu_count" comment:"菜单数量"`
	DeptCount   int64             `json:"dept_count" comment:"部门数量"`
	OnlineCount int64             `json:"online_count" comment:"在线用户数量"`
	LoginTrend  []TrendItem       `json:"login_trend" comment:"登录趋势"`
	RecentLogin []LoginRecordItem `json:"recent_login" comment:"最近登录记录"`
	DeptUsers   []DeptUserItem    `json:"dept_users" comment:"部门用户统计"`
}
