package model

// SysStorageConfig 文件存储渠道配置。
// Params 保存渠道参数 JSON，不同渠道字段不同（见前端 enums/storage.ts 的字段模板）。
// is_default 标记当前生效的默认渠道，运行时通过 common/upload 动态读取。表：sys_storage_config
type SysStorageConfig struct {
	Base
	Name      string `gorm:"size:64;not null;comment:渠道名称" json:"name"`
	Channel   string `gorm:"size:32;not null;comment:存储渠道" json:"channel"`
	Params    string `gorm:"type:text;comment:渠道参数JSON" json:"params"`
	IsDefault int    `gorm:"default:0;comment:是否默认渠道" json:"is_default"`
	Status    int    `gorm:"default:1;comment:状态" json:"status"`
	Sort      int    `gorm:"default:0;comment:排序" json:"sort"`
	Remark    string `gorm:"size:255;default:'';comment:备注" json:"remark"`
}

func (SysStorageConfig) TableName() string { return "sys_storage_config" }
