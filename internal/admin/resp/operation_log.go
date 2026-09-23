package resp

type OperationLogListRes struct {
	List  []OperationLogItem `json:"list" comment:"操作日志列表"`
	Total int64              `json:"total" comment:"数据总数"`
}
