package param

// CategorySaveReq 新闻分类保存请求。
type CategorySaveReq struct {
	ID     uint   `json:"id" comment:"主键ID，新增时为0"`
	Name   string `json:"name" binding:"required,max=64" comment:"分类名称"`
	Slug   string `json:"slug" binding:"required,max=64,alphanum" comment:"分类标识"`
	Sort   int    `json:"sort" binding:"min=0" comment:"排序值"`
	Status int    `json:"status" binding:"required,oneof=1 2" comment:"状态，1启用，2停用"`
	Remark string `json:"remark" binding:"max=255" comment:"分类备注"`
}

// CategoryIDReq 新闻分类ID请求。
type CategoryIDReq struct {
	ID uint `json:"id" form:"id" binding:"required" comment:"分类ID"`
}
