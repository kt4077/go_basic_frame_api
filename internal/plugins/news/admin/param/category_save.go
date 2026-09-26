package param

// CategorySaveReq 新闻分类保存请求。
type CategorySaveReq struct {
	ID     uint   `json:"id" validate:"主键ID" comment:"主键ID，新增时为0"`
	Name   string `json:"name" binding:"required,max=64" validate:"分类名称" comment:"分类名称"`
	Slug   string `json:"slug" binding:"required,max=64,alphanum" validate:"分类标识" comment:"分类标识"`
	Sort   int    `json:"sort" binding:"min=0" validate:"排序值" comment:"排序值"`
	Status int    `json:"status" binding:"required,oneof=1 2" validate:"状态" comment:"状态，1启用，2停用"`
	Remark string `json:"remark" binding:"max=255" validate:"分类备注" comment:"分类备注"`
}

// CategoryIDReq 新闻分类ID请求。
type CategoryIDReq struct {
	ID uint `json:"id" form:"id" binding:"required" validate:"分类ID" comment:"分类ID"`
}
