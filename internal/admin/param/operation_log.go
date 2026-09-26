package param

type OperationLogListReq struct {
	Username  string `form:"username" validate:"操作账号" comment:"操作账号"`
	StartTime string `form:"start_time" validate:"开始时间" comment:"开始时间"`
	EndTime   string `form:"end_time" validate:"结束时间" comment:"结束时间"`
	Page      int    `form:"page" validate:"页码" comment:"页码"`
	PageSize  int    `form:"page_size" validate:"每页数量" comment:"每页数量"`
}

// OperationLogDeleteReq 批量删除操作日志请求。
type OperationLogDeleteReq struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=500,dive,gt=0" validate:"日志ID列表" comment:"需要物理删除的日志ID列表，最多500条"`
}

// OperationLogClearReq 全量清空操作日志请求。
type OperationLogClearReq struct {
	Confirm string `json:"confirm" binding:"required,eq=CLEAR_ALL_OPERATION_LOGS" validate:"全量清空确认标识" comment:"全量清空确认标识"`
}
