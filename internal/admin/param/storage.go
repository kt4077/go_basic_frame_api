package param

type StorageSaveReq struct {
	ID        uint   `json:"id" comment:"主键ID"`
	Name      string `json:"name" binding:"required" comment:"渠道名称"`
	Channel   string `json:"channel" binding:"required" comment:"存储渠道"`
	Params    string `json:"params" comment:"渠道参数JSON"`
	IsDefault int    `json:"is_default" comment:"是否默认"`
	Status    int    `json:"status" comment:"状态"`
	Sort      int    `json:"sort" comment:"排序"`
	Remark    string `json:"remark" comment:"备注"`
}
