package param

type OperationLogListReq struct {
	Username  string `form:"username" comment:"操作账号"`
	StartTime string `form:"start_time" comment:"开始时间"`
	EndTime   string `form:"end_time" comment:"结束时间"`
	Page      int    `form:"page" comment:"页码"`
	PageSize  int    `form:"page_size" comment:"每页数量"`
}
