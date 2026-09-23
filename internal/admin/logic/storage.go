package logic

import (
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/dberror"
)

type StorageLogic struct{ App *app.App }

// List 存储渠道列表（配置项数量少，不分页，按排序返回）。
func (l *StorageLogic) List(c *gin.Context) ([]resp.StorageItem, error) {
	var list []model.SysStorageConfig
	if err := l.App.DB.Order("sort ASC, id ASC").Find(&list).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return resp.NewStorageItems(list), nil
}

// Create 新增存储渠道。第一个渠道自动设为默认。
func (l *StorageLogic) Create(c *gin.Context, req *param.StorageSaveReq) (*resp.StorageItem, error) {
	if err := validateStorageParams(req.Params); err != nil {
		return nil, err
	}
	if req.IsDefault == enums.StorageDefaultYes && req.Status != enums.StatusEnabled {
		return nil, errors.New("禁用状态的渠道不能设为默认")
	}
	cfg := model.SysStorageConfig{
		Name: req.Name, Channel: req.Channel, Params: req.Params,
		IsDefault: req.IsDefault, Status: req.Status, Sort: req.Sort, Remark: req.Remark,
	}
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.SysStorageConfig{}).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if cfg.Status != enums.StatusEnabled {
				return errors.New("首个存储渠道必须启用")
			}
			cfg.IsDefault = enums.StorageDefaultYes
		}
		if err := tx.Create(&cfg).Error; err != nil {
			return err
		}
		return clearOtherDefaults(tx, cfg.ID, cfg.IsDefault)
	})
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("默认渠道已被其他请求更新，请刷新后重试")
		}
		if err.Error() == "首个存储渠道必须启用" {
			return nil, err
		}
		return nil, errors.New("创建失败")
	}
	result := resp.NewStorageItem(cfg)
	return &result, nil
}

// Update 修改存储渠道。
func (l *StorageLogic) Update(c *gin.Context, req *param.StorageSaveReq) (*resp.StorageItem, error) {
	var cfg model.SysStorageConfig
	if err := l.App.DB.First(&cfg, req.ID).Error; err != nil {
		return nil, errors.New("存储渠道不存在")
	}
	if err := validateStorageParams(req.Params); err != nil {
		return nil, err
	}
	if req.IsDefault == enums.StorageDefaultYes && req.Status != enums.StatusEnabled {
		return nil, errors.New("禁用状态的渠道不能设为默认")
	}
	if cfg.IsDefault == enums.StorageDefaultYes &&
		(req.IsDefault != enums.StorageDefaultYes || req.Status != enums.StatusEnabled) {
		return nil, errors.New("默认渠道不能取消默认或禁用，请先将其他渠道设为默认")
	}
	cfg.Name, cfg.Channel, cfg.Params = req.Name, req.Channel, req.Params
	cfg.IsDefault, cfg.Status, cfg.Sort, cfg.Remark = req.IsDefault, req.Status, req.Sort, req.Remark
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&cfg).Error; err != nil {
			return err
		}
		return clearOtherDefaults(tx, cfg.ID, cfg.IsDefault)
	})
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("默认渠道已被其他请求更新，请刷新后重试")
		}
		return nil, errors.New("修改失败")
	}
	result := resp.NewStorageItem(cfg)
	return &result, nil
}

// SetDefault 设为默认渠道。
func (l *StorageLogic) SetDefault(c *gin.Context, req *param.IDReq) error {
	var cfg model.SysStorageConfig
	if err := l.App.DB.First(&cfg, req.ID).Error; err != nil {
		return errors.New("存储渠道不存在")
	}
	if cfg.Status != enums.StatusEnabled {
		return errors.New("禁用状态的渠道不能设为默认")
	}
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.SysStorageConfig{}).Where("1 = 1").
			Update("is_default", enums.StorageDefaultNo).Error; err != nil {
			return err
		}
		return tx.Model(&cfg).Update("is_default", enums.StorageDefaultYes).Error
	})
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return errors.New("默认渠道已被其他请求更新，请刷新后重试")
		}
		return errors.New("设置失败")
	}
	return nil
}

// Delete 删除存储渠道。默认渠道不允许删除，需先切换默认。
func (l *StorageLogic) Delete(c *gin.Context, req *param.IDReq) error {
	var cfg model.SysStorageConfig
	if err := l.App.DB.First(&cfg, req.ID).Error; err != nil {
		return errors.New("存储渠道不存在")
	}
	if cfg.IsDefault == enums.StorageDefaultYes {
		return errors.New("默认渠道不允许删除，请先将其他渠道设为默认")
	}
	if err := l.App.DB.Delete(&cfg).Error; err != nil {
		return errors.New("删除失败")
	}
	return nil
}

// clearOtherDefaults 保证全局只有一个默认渠道。
func clearOtherDefaults(tx *gorm.DB, id uint, isDefault int) error {
	if isDefault != enums.StorageDefaultYes {
		return nil
	}
	return tx.Model(&model.SysStorageConfig{}).
		Where("id != ? AND is_default = ?", id, enums.StorageDefaultYes).
		Update("is_default", enums.StorageDefaultNo).Error
}

// validateStorageParams 渠道参数必须是合法的 JSON 对象。
func validateStorageParams(params string) error {
	if params == "" {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(params), &m); err != nil {
		return errors.New("渠道参数必须是合法的 JSON 对象")
	}
	return nil
}
