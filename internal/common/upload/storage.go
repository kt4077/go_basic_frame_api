// 存储渠道的运行时读取：每次按需从数据库取当前默认渠道配置，
// 管理端修改后无需重启即可生效（动态配置）。
package upload

import (
	"encoding/json"
	"errors"

	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/oss"
)

// LoadDefaultStorage 读取当前启用中的默认存储渠道配置。
// 业务代码需要上传文件时调用本方法获取最新配置，而不是在启动时缓存。
func LoadDefaultStorage(application *app.App) (*model.SysStorageConfig, error) {
	var cfg model.SysStorageConfig
	err := application.DB.
		Where("is_default = ? AND status = ?", enums.StorageDefaultYes, enums.StatusEnabled).
		Order("id ASC").
		First(&cfg).Error
	if err != nil {
		return nil, errors.New("尚未配置可用的默认存储渠道，请到 系统配置/渠道配置/存储配置 中设置")
	}
	return &cfg, nil
}

// StorageParams 将渠道参数 JSON 解析为 map，空串返回空 map。
func StorageParams(cfg *model.SysStorageConfig) (map[string]string, error) {
	out := map[string]string{}
	if cfg.Params == "" {
		return out, nil
	}
	if err := json.Unmarshal([]byte(cfg.Params), &out); err != nil {
		return nil, errors.New("存储渠道参数不是合法的 JSON")
	}
	return out, nil
}

// LoadUploader 按当前默认存储渠道构造对应平台的上传器（动态配置，随管理端修改实时生效）。
func LoadUploader(application *app.App) (oss.Uploader, error) {
	cfg, err := LoadDefaultStorage(application)
	if err != nil {
		return nil, err
	}
	params, err := StorageParams(cfg)
	if err != nil {
		return nil, err
	}
	return oss.NewUploader(cfg.Channel, params)
}
