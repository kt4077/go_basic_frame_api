package plugin

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"sync"

	"gorm.io/gorm"

	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
)

var pluginIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)

// Registry 管理编译进程序的插件、启用状态、路由与生命周期。
type Registry struct {
	app      *app.App
	plugins  map[string]Plugin
	enabled  []Plugin
	started  []Plugin
	prepared bool
	mu       sync.Mutex
}

// NewRegistry 创建空插件注册中心。
func NewRegistry(application *app.App) *Registry {
	return &Registry{app: application, plugins: make(map[string]Plugin)}
}

// Register 注册一个编译期插件，同一 PluginID 不允许重复。
func (r *Registry) Register(item Plugin) error {
	if item == nil {
		return errors.New("不能注册空插件")
	}
	manifest := item.Manifest()
	if err := validateManifest(manifest); err != nil {
		return err
	}
	if _, exists := r.plugins[manifest.PluginID]; exists {
		return fmt.Errorf("插件 %s 重复注册", manifest.PluginID)
	}
	r.plugins[manifest.PluginID] = item
	return nil
}

// Manifest 返回编译进当前程序的插件清单。
func (r *Registry) Manifest(pluginID string) (Manifest, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, exists := r.plugins[pluginID]
	if !exists {
		return Manifest{}, false
	}
	return item.Manifest(), true
}

// ValidateEnable 校验插件是否具备下次重启时启用的条件。
func (r *Registry) ValidateEnable(ctx context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, exists := r.plugins[pluginID]
	if !exists {
		return fmt.Errorf("插件 %s 未编译进当前程序", pluginID)
	}
	var record model.SysPlugin
	if err := r.app.DB.WithContext(ctx).Where("plugin_id = ?", pluginID).First(&record).Error; err != nil {
		return errors.New("插件记录不存在")
	}
	manifest := item.Manifest()
	if record.Version != manifest.Version {
		return fmt.Errorf("数据库版本为 %s，程序版本为 %s，请先完成升级", record.Version, manifest.Version)
	}
	if err := r.checkMigrations(ctx, item); err != nil {
		return err
	}
	for _, dependency := range manifest.Dependencies {
		var dependencyRecord model.SysPlugin
		if err := r.app.DB.WithContext(ctx).
			Where("plugin_id = ? AND status = ?", dependency.PluginID, enums.PluginStatusEnabled).
			First(&dependencyRecord).Error; err != nil {
			return fmt.Errorf("缺少已启用依赖 %s", dependency.PluginID)
		}
		dependencyPlugin, compiled := r.plugins[dependency.PluginID]
		if !compiled {
			return fmt.Errorf("依赖插件 %s 未编译进当前程序", dependency.PluginID)
		}
		if dependency.Version != "" && dependencyPlugin.Manifest().Version != dependency.Version {
			return fmt.Errorf("要求依赖 %s 版本 %s，当前为 %s", dependency.PluginID, dependency.Version, dependencyPlugin.Manifest().Version)
		}
	}
	return nil
}

// ValidateDisable 校验是否存在仍在使用当前插件的已启用插件。
func (r *Registry) ValidateDisable(ctx context.Context, pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var records []model.SysPlugin
	if err := r.app.DB.WithContext(ctx).Where("status = ?", enums.PluginStatusEnabled).Find(&records).Error; err != nil {
		return errors.New("检查插件依赖失败")
	}
	for _, record := range records {
		if record.PluginID == pluginID {
			continue
		}
		item, exists := r.plugins[record.PluginID]
		if !exists {
			continue
		}
		for _, dependency := range item.Manifest().Dependencies {
			if dependency.PluginID == pluginID {
				return fmt.Errorf("插件 %s 正在依赖当前插件，不能停用", record.Name)
			}
		}
	}
	return nil
}

// Prepare 读取数据库启用状态并校验插件版本、依赖和迁移记录。
// 未安装或已停用的编译期插件不会注册路由或启动任务。
func (r *Registry) Prepare(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.prepared {
		return nil
	}

	var records []model.SysPlugin
	if err := r.app.DB.WithContext(ctx).Where("status = ?", enums.PluginStatusEnabled).Find(&records).Error; err != nil {
		return fmt.Errorf("读取启用插件失败: %w", err)
	}

	enabledByID := make(map[string]Plugin, len(records))
	for _, record := range records {
		item, exists := r.plugins[record.PluginID]
		if !exists {
			return fmt.Errorf("插件 %s 已启用，但当前程序未编译该插件", record.PluginID)
		}
		manifest := item.Manifest()
		if record.Version != manifest.Version {
			return fmt.Errorf("插件 %s 数据库版本为 %s，程序版本为 %s，请先完成升级", record.PluginID, record.Version, manifest.Version)
		}
		if err := r.checkMigrations(ctx, item); err != nil {
			return err
		}
		enabledByID[record.PluginID] = item
	}

	for pluginID, item := range enabledByID {
		for _, dependency := range item.Manifest().Dependencies {
			dependencyPlugin, exists := enabledByID[dependency.PluginID]
			if !exists {
				return fmt.Errorf("插件 %s 缺少已启用依赖 %s", pluginID, dependency.PluginID)
			}
			if dependency.Version != "" && dependencyPlugin.Manifest().Version != dependency.Version {
				return fmt.Errorf("插件 %s 要求依赖 %s 版本 %s，当前为 %s", pluginID, dependency.PluginID, dependency.Version, dependencyPlugin.Manifest().Version)
			}
		}
	}

	ordered, err := sortByDependencies(enabledByID)
	if err != nil {
		return err
	}
	r.enabled = ordered
	r.prepared = true
	return nil
}

// RegisterAdminRoutes 将已启用插件注册到管理端受控路由组。
func (r *Registry) RegisterAdminRoutes(groups AdminRouteGroups) error {
	if !r.prepared {
		return errors.New("插件注册中心尚未准备完成")
	}
	pluginContext := &Context{App: r.app}
	for _, item := range r.enabled {
		if err := item.RegisterAdminRoutes(pluginContext, groups); err != nil {
			return fmt.Errorf("插件 %s 注册管理端路由失败: %w", item.Manifest().PluginID, err)
		}
	}
	return nil
}

// RegisterAPIRoutes 将已启用插件注册到用户端受控路由组。
func (r *Registry) RegisterAPIRoutes(groups APIRouteGroups) error {
	if !r.prepared {
		return errors.New("插件注册中心尚未准备完成")
	}
	pluginContext := &Context{App: r.app}
	for _, item := range r.enabled {
		if err := item.RegisterAPIRoutes(pluginContext, groups); err != nil {
			return fmt.Errorf("插件 %s 注册用户端路由失败: %w", item.Manifest().PluginID, err)
		}
	}
	return nil
}

// Start 按插件 ID 顺序启动已启用插件的后台能力。
func (r *Registry) Start(ctx context.Context, service ServiceType) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, item := range r.enabled {
		if err := item.Start(ctx, service); err != nil {
			for index := len(r.started) - 1; index >= 0; index-- {
				_ = r.started[index].Stop(ctx)
			}
			r.started = nil
			return fmt.Errorf("插件 %s 启动失败: %w", item.Manifest().PluginID, err)
		}
		r.started = append(r.started, item)
	}
	return nil
}

// Stop 按启动相反顺序停止插件，返回遇到的第一个错误。
func (r *Registry) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	var firstErr error
	for index := len(r.started) - 1; index >= 0; index-- {
		item := r.started[index]
		if err := item.Stop(ctx); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("插件 %s 停止失败: %w", item.Manifest().PluginID, err)
		}
	}
	r.started = nil
	return firstErr
}

func (r *Registry) checkMigrations(ctx context.Context, item Plugin) error {
	pluginID := item.Manifest().PluginID
	for _, migration := range item.Migrations() {
		if migration.Version == "" {
			return fmt.Errorf("插件 %s 存在空迁移版本", pluginID)
		}
		var record model.SysPluginMigration
		err := r.app.DB.WithContext(ctx).
			Where("plugin_id = ? AND version = ? AND status = ?", pluginID, migration.Version, enums.PluginMigrationSuccess).
			First(&record).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("插件 %s 缺少迁移 %s，请先执行插件升级 SQL", pluginID, migration.Version)
		}
		if err != nil {
			return fmt.Errorf("检查插件 %s 迁移 %s 失败: %w", pluginID, migration.Version, err)
		}
		if migration.Checksum != "" && record.Checksum != migration.Checksum {
			return fmt.Errorf("插件 %s 迁移 %s 摘要不一致", pluginID, migration.Version)
		}
	}
	return nil
}

func validateManifest(manifest Manifest) error {
	if !pluginIDPattern.MatchString(manifest.PluginID) {
		return fmt.Errorf("插件标识 %q 不合法，只允许小写字母、数字和下划线，且必须以字母开头", manifest.PluginID)
	}
	if manifest.Name == "" {
		return fmt.Errorf("插件 %s 缺少名称", manifest.PluginID)
	}
	if manifest.Version == "" {
		return fmt.Errorf("插件 %s 缺少版本", manifest.PluginID)
	}
	if manifest.CoreVersion != "" && manifest.CoreVersion != CoreVersion {
		return fmt.Errorf("插件 %s 要求核心版本 %s，当前核心版本为 %s", manifest.PluginID, manifest.CoreVersion, CoreVersion)
	}
	for _, dependency := range manifest.Dependencies {
		if !pluginIDPattern.MatchString(dependency.PluginID) {
			return fmt.Errorf("插件 %s 的依赖标识 %q 不合法", manifest.PluginID, dependency.PluginID)
		}
		if dependency.PluginID == manifest.PluginID {
			return fmt.Errorf("插件 %s 不能依赖自身", manifest.PluginID)
		}
	}
	return nil
}

func sortByDependencies(enabled map[string]Plugin) ([]Plugin, error) {
	ids := make([]string, 0, len(enabled))
	for pluginID := range enabled {
		ids = append(ids, pluginID)
	}
	sort.Strings(ids)

	states := make(map[string]int, len(enabled))
	ordered := make([]Plugin, 0, len(enabled))
	var visit func(string) error
	visit = func(pluginID string) error {
		switch states[pluginID] {
		case 1:
			return fmt.Errorf("插件依赖存在循环，涉及插件 %s", pluginID)
		case 2:
			return nil
		}
		states[pluginID] = 1
		item := enabled[pluginID]
		for _, dependency := range item.Manifest().Dependencies {
			if err := visit(dependency.PluginID); err != nil {
				return err
			}
		}
		states[pluginID] = 2
		ordered = append(ordered, item)
		return nil
	}

	for _, pluginID := range ids {
		if err := visit(pluginID); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}
