package logic

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/internal/common/upload"
)

// PlatformLogic 平台配置业务逻辑。
type PlatformLogic struct{ App *app.App }

// AdminDetail 查询管理端平台配置。
func (l *PlatformLogic) AdminDetail(c *gin.Context) (*resp.AdminPlatformConfigRes, error) {
	var config model.SysPlatformConfig
	err := l.App.DB.Where("type = ?", enums.PlatformTypeAdmin).First(&config).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("查询管理端配置失败")
	}
	logo, err := upload.FileURL(l.App, config.Logo)
	if err != nil {
		return nil, errors.New("管理端Logo地址解析失败")
	}
	return &resp.AdminPlatformConfigRes{Logo: logo, LogoPath: config.Logo, SystemName: config.SystemName, Version: l.App.Cfg.Version}, nil
}

// SaveAdmin 保存管理端平台配置。
func (l *PlatformLogic) SaveAdmin(c *gin.Context, req *param.AdminPlatformSaveReq) (*resp.AdminPlatformConfigRes, error) {
	logo, err := upload.NormalizeFilePath(l.App, req.Logo)
	if err != nil {
		return nil, err
	}
	config := model.SysPlatformConfig{Type: enums.PlatformTypeAdmin, Logo: logo, SystemName: req.SystemName}
	if err := l.upsert(&config, []string{"logo", "system_name"}); err != nil {
		return nil, errors.New("保存管理端配置失败")
	}
	return l.AdminDetail(c)
}

// UserDetail 查询用户端平台配置。
func (l *PlatformLogic) UserDetail(c *gin.Context) (*resp.UserPlatformConfigRes, error) {
	var config model.SysPlatformConfig
	err := l.App.DB.Where("type = ?", enums.PlatformTypeAPI).First(&config).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("查询用户端配置失败")
	}
	avatar, err := upload.FileURL(l.App, config.DefaultAvatar)
	if err != nil {
		return nil, errors.New("默认头像地址解析失败")
	}
	return &resp.UserPlatformConfigRes{DefaultNickname: config.DefaultNickname, DefaultAvatar: avatar, DefaultAvatarPath: config.DefaultAvatar}, nil
}

// SaveUser 保存用户端平台配置。
func (l *PlatformLogic) SaveUser(c *gin.Context, req *param.UserPlatformSaveReq) (*resp.UserPlatformConfigRes, error) {
	avatar, err := upload.NormalizeFilePath(l.App, req.DefaultAvatar)
	if err != nil {
		return nil, err
	}
	config := model.SysPlatformConfig{Type: enums.PlatformTypeAPI, DefaultNickname: req.DefaultNickname, DefaultAvatar: avatar}
	if err := l.upsert(&config, []string{"default_nickname", "default_avatar"}); err != nil {
		return nil, errors.New("保存用户端配置失败")
	}
	return l.UserDetail(c)
}

// upsert 根据平台类型原子新增或更新，唯一索引保证并发保存不会产生重复配置。
func (l *PlatformLogic) upsert(config *model.SysPlatformConfig, columns []string) error {
	columns = append(columns, "updated_at")
	return l.App.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}},
		DoUpdates: clause.AssignmentColumns(columns),
	}).Create(config).Error
}
