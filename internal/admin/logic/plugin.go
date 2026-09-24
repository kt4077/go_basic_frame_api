package logic

import (
	"errors"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	commonplugin "server_api/internal/common/plugin"
	"server_api/internal/common/upload"
	"server_api/pkg/pagination"
)

// PluginLogic 插件管理业务逻辑。
type PluginLogic struct {
	App      *app.App
	Registry *commonplugin.Registry
}

// List 返回已安装插件分页列表，并标记当前程序是否已编译对应插件。
func (l *PluginLogic) List(c *gin.Context, req *param.PluginListReq) (*resp.PluginListRes, error) {
	var total int64
	var records []model.SysPlugin
	db := l.App.DB.WithContext(c.Request.Context()).Model(&model.SysPlugin{})
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("plugin_id LIKE ? OR name LIKE ?", keyword, keyword)
	}
	if req.Status != 0 {
		db = db.Where("status = ?", req.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询插件数量失败")
	}
	page, pageSize := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error; err != nil {
		return nil, errors.New("查询插件列表失败")
	}
	items := make([]resp.PluginItem, 0, len(records))
	for _, record := range records {
		item, err := l.newPluginItem(record)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return &resp.PluginListRes{List: items, Total: total}, nil
}

// Detail 返回插件基础信息、编译状态和迁移记录。
func (l *PluginLogic) Detail(c *gin.Context, req *param.PluginIDReq) (*resp.PluginDetail, error) {
	var record model.SysPlugin
	if err := l.App.DB.WithContext(c.Request.Context()).Where("plugin_id = ?", req.PluginID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("插件不存在")
		}
		return nil, errors.New("查询插件详情失败")
	}
	var migrationRecords []model.SysPluginMigration
	if err := l.App.DB.WithContext(c.Request.Context()).Where("plugin_id = ?", req.PluginID).
		Order("executed_at DESC, id DESC").Find(&migrationRecords).Error; err != nil {
		return nil, errors.New("查询插件迁移记录失败")
	}
	migrations := make([]resp.PluginMigrationItem, 0, len(migrationRecords))
	for _, migration := range migrationRecords {
		migrations = append(migrations, resp.PluginMigrationItem{
			ID: migration.ID, Version: migration.Version, Checksum: migration.Checksum,
			Status: migration.Status, ExecutionMs: migration.ExecutionMs,
			ErrorMessage: migration.ErrorMessage, ExecutedAt: migration.ExecutedAt,
		})
	}
	item, err := l.newPluginItem(record)
	if err != nil {
		return nil, err
	}
	return &resp.PluginDetail{PluginItem: item, Manifest: record.Manifest, Migrations: migrations}, nil
}

// UpdateStatus 修改插件状态。状态只写入数据库，路由和后台任务在服务重启后生效。
func (l *PluginLogic) UpdateStatus(c *gin.Context, req *param.PluginStatusReq) error {
	ctx := c.Request.Context()
	var record model.SysPlugin
	if err := l.App.DB.WithContext(ctx).Where("plugin_id = ?", req.PluginID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("插件不存在")
		}
		return errors.New("查询插件状态失败")
	}
	if record.Status == req.Status {
		return nil
	}
	if req.Status == enums.PluginStatusEnabled {
		if err := l.Registry.ValidateEnable(ctx, req.PluginID); err != nil {
			return err
		}
	} else {
		if err := l.Registry.ValidateDisable(ctx, req.PluginID); err != nil {
			return err
		}
	}
	result := l.App.DB.WithContext(ctx).Model(&model.SysPlugin{}).
		Where("plugin_id = ? AND status = ?", req.PluginID, record.Status).
		Update("status", req.Status)
	if result.Error != nil {
		return errors.New("修改插件状态失败")
	}
	if result.RowsAffected == 0 {
		return errors.New("插件状态已被其他请求修改，请刷新后重试")
	}
	return nil
}

// UpdateInfo 修改插件作者、主页和描述，不允许修改运行标识及版本信息。
func (l *PluginLogic) UpdateInfo(c *gin.Context, req *param.PluginInfoUpdateReq) error {
	author := strings.TrimSpace(req.Author)
	homepage := strings.TrimSpace(req.Homepage)
	description := strings.TrimSpace(req.Description)
	logo, err := normalizeFilePath(l.App, req.Logo)
	if err != nil {
		return errors.New("插件Logo路径不合法")
	}
	if len(logo) > 500 {
		return errors.New("插件Logo相对路径不能超过500个字符")
	}
	if homepage != "" {
		parsedURL, err := url.Parse(homepage)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
			return errors.New("插件地址必须是有效的 HTTP 或 HTTPS 地址")
		}
	}
	var count int64
	if err := l.App.DB.WithContext(c.Request.Context()).Model(&model.SysPlugin{}).
		Where("plugin_id = ?", req.PluginID).Count(&count).Error; err != nil {
		return errors.New("查询插件信息失败")
	}
	if count == 0 {
		return errors.New("插件不存在")
	}
	result := l.App.DB.WithContext(c.Request.Context()).Model(&model.SysPlugin{}).
		Where("plugin_id = ?", req.PluginID).
		Updates(map[string]interface{}{
			"logo": logo, "author": author, "homepage": homepage, "description": description,
		})
	if result.Error != nil {
		return errors.New("修改插件信息失败")
	}
	return nil
}

func (l *PluginLogic) newPluginItem(record model.SysPlugin) (resp.PluginItem, error) {
	manifest, compiled := l.Registry.Manifest(record.PluginID)
	codeVersion := ""
	if compiled {
		codeVersion = manifest.Version
	}
	logoURL, err := upload.FileURL(l.App, record.Logo)
	if err != nil {
		return resp.PluginItem{}, errors.New("插件Logo地址解析失败")
	}
	return resp.PluginItem{
		BaseItem: resp.BaseItem{
			ID: record.ID, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
			DeletedAt: record.DeletedAt,
		},
		PluginID: record.PluginID, Name: record.Name, Version: record.Version,
		Logo: record.Logo, LogoURL: logoURL,
		CodeVersion: codeVersion, Author: record.Author, Homepage: record.Homepage,
		Description: record.Description, Status: record.Status, Compiled: compiled,
	}, nil
}
