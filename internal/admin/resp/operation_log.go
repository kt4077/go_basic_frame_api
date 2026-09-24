package resp

type OperationLogListRes struct {
	List  []OperationLogItem `json:"list" comment:"操作日志列表"`
	Total int64              `json:"total" comment:"数据总数"`
}

// OperationLogDeleteRes 操作日志物理删除结果。
type OperationLogDeleteRes struct {
	Deleted int64 `json:"deleted" comment:"实际删除数量"`
}
