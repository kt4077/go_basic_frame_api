package model

import "time"

// SysOperationLog 用户操作日志表。管理端除登录外的所有接口调用均自动记录，
// 含请求参数（脱敏）与接口响应内容。表：sys_operation_log
type SysOperationLog struct {
	ID             uint      `gorm:"primaryKey;comment:主键ID" json:"id"`
	CreatedAt      time.Time `gorm:"comment:操作时间" json:"created_at"`
	UpdatedAt      time.Time `gorm:"comment:更新时间" json:"updated_at"`
	UserID         *uint     `gorm:"index;comment:操作人ID" json:"user_id"`
	Username       string    `gorm:"size:32;default:'';comment:操作人账号" json:"username"`
	Method         string    `gorm:"size:16;default:'';comment:请求方式" json:"method"`
	Path           string    `gorm:"size:128;default:'';comment:请求路由" json:"path"`
	RequestParams  string    `gorm:"type:text;comment:脱敏后的请求参数" json:"request_params"`
	ResponseParams string    `gorm:"type:mediumtext;comment:接口响应内容" json:"response_params"`
	IP             string    `gorm:"size:64;default:'';comment:客户端IP" json:"ip"`
	UserAgent      string    `gorm:"size:255;default:'';comment:用户代理" json:"user_agent"`
	Code           int       `gorm:"not null;default:0;comment:业务响应码" json:"code"`
	CostMs         int64     `gorm:"not null;default:0;comment:耗时毫秒" json:"cost_ms"`
	Client         string    `gorm:"size:16;default:'';comment:客户端类型" json:"client"`
}

func (SysOperationLog) TableName() string { return "sys_operation_log" }
