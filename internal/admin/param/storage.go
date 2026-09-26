package param

type StorageSaveReq struct {
	ID        uint   `json:"id" validate:"主键ID" comment:"主键ID"`
	Name      string `json:"name" binding:"required" validate:"渠道名称" comment:"渠道名称"`
	Channel   string `json:"channel" binding:"required" validate:"存储渠道" comment:"存储渠道"`
	Params    string `json:"params" validate:"渠道参数JSON" comment:"渠道参数JSON"`
	IsDefault int    `json:"is_default" validate:"是否默认" comment:"是否默认"`
	Status    int    `json:"status" validate:"状态" comment:"状态"`
	Sort      int    `json:"sort" validate:"排序" comment:"排序"`
	Remark    string `json:"remark" validate:"备注" comment:"备注"`
}
