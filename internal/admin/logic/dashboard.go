package logic

import (
	"time"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
)

type DashboardLogic struct{ App *app.App }

// Overview 系统总览：核心数量统计 + 近7天登录趋势 + 最近登录记录 + 各部门人数。
func (l *DashboardLogic) Overview(c *gin.Context) (*resp.OverviewRes, error) {
	var userCount, roleCount, menuCount, deptCount, onlineCount int64
	l.App.DB.Model(&model.SysUser{}).Count(&userCount)
	l.App.DB.Model(&model.SysRole{}).Count(&roleCount)
	l.App.DB.Model(&model.SysMenu{}).Count(&menuCount)
	l.App.DB.Model(&model.SysDept{}).Count(&deptCount)
	l.App.DB.Model(&model.SysUserLogin{}).
		Where("status = ?", enums.LoginStatusOnline).Count(&onlineCount)

	// 近7天登录趋势（按天统计登录次数）
	trend := make([]resp.TrendItem, 0, 7)
	start := time.Now().AddDate(0, 0, -6)
	for i := 0; i < 7; i++ {
		day := start.AddDate(0, 0, i)
		dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
		var n int64
		l.App.DB.Model(&model.SysUserLogin{}).
			Where("login_at >= ? AND login_at < ?", dayStart, dayStart.Add(24*time.Hour)).Count(&n)
		trend = append(trend, resp.TrendItem{Name: day.Format("01-02"), Value: n})
	}

	// 最近10条登录记录
	var recent []model.SysUserLogin
	l.App.DB.Order("login_at DESC").Limit(10).Find(&recent)

	// 各部门人数
	var deptUsers []resp.DeptUserItem
	l.App.DB.Model(&model.SysUser{}).
		Select("sys_user.dept_id", "IFNULL(sys_dept.name, '未分配') AS name", "COUNT(*) AS total").
		Joins("LEFT JOIN sys_dept ON sys_dept.id = sys_user.dept_id").
		Group("sys_user.dept_id, sys_dept.name").
		Scan(&deptUsers)

	return &resp.OverviewRes{
		UserCount:   userCount,
		RoleCount:   roleCount,
		MenuCount:   menuCount,
		DeptCount:   deptCount,
		OnlineCount: onlineCount,
		LoginTrend:  trend,
		RecentLogin: resp.NewLoginRecordItems(recent),
		DeptUsers:   deptUsers,
	}, nil
}
