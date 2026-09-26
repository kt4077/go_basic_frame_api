package logic

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/model"
	"server_api/pkg/pagination"
)

type OperationLogLogic struct{ App *app.App }

// List 操作日志分页列表，按操作时间倒序，支持按操作人过滤。
func (l *OperationLogLogic) List(c *gin.Context, req *param.OperationLogListReq) (*resp.OperationLogListRes, error) {
	var total int64
	var list []model.SysOperationLog
	db := l.App.DB.Model(&model.SysOperationLog{})
	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.StartTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", req.StartTime, time.Local)
		if err != nil {
			return nil, errors.New("开始时间格式错误")
		}
		db = db.Where("created_at >= ?", t)
	}
	if req.EndTime != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", req.EndTime, time.Local)
		if err != nil {
			return nil, errors.New("结束时间格式错误")
		}
		db = db.Where("created_at <= ?", t)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	page, pageSize := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Order("created_at DESC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return &resp.OperationLogListRes{List: resp.NewOperationLogItems(list), Total: total}, nil
}

// Delete 批量物理删除指定操作日志。
func (l *OperationLogLogic) Delete(c *gin.Context, req *param.OperationLogDeleteReq) (*resp.OperationLogDeleteRes, error) {
	result := l.App.DB.WithContext(c).Unscoped().
		Where("id IN ?", req.IDs).
		Delete(&model.SysOperationLog{})
	if result.Error != nil {
		return nil, errors.New("删除操作日志失败")
	}
	return &resp.OperationLogDeleteRes{Deleted: result.RowsAffected}, nil
}

// Clear 使用单条 DELETE 语句全量物理删除操作日志，保证该次清理的原子性。
func (l *OperationLogLogic) Clear(c *gin.Context, req *param.OperationLogClearReq) (*resp.OperationLogDeleteRes, error) {
	if req.Confirm != "CLEAR_ALL_OPERATION_LOGS" {
		return nil, errors.New("清空确认信息错误")
	}
	result := l.App.DB.WithContext(c).Exec("DELETE FROM `sys_operation_log`")
	if result.Error != nil {
		return nil, errors.New("清空操作日志失败")
	}
	return &resp.OperationLogDeleteRes{Deleted: result.RowsAffected}, nil
}
