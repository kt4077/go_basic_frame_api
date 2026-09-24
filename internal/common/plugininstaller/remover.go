package plugininstaller

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"server_api/config"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
)

// RemoveOptions 插件移除与彻底清理参数。
type RemoveOptions struct {
	ServerRoot string
	AdminRoot  string
	ConfigPath string
}

// RemoveResult 插件移除结果。
type RemoveResult struct {
	PluginID      string
	MenusRemoved  int64
	ServerRemoved bool
	AdminRemoved  bool
}

// PurgeResult 插件彻底清理结果。
type PurgeResult struct {
	RemoveResult
	DroppedTables []string
}

// Remove 非破坏性移除插件运行能力，保留业务表、迁移历史、安装日志和上传文件。
// 数据库元数据先原子移除；随后清理源码。若进程在两阶段之间退出，插件因数据库记录
// 已移除而不会运行，重复执行命令可继续完成源码清理。
func Remove(pluginID string, options RemoveOptions) (*RemoveResult, error) {
	if !keyPattern.MatchString(pluginID) {
		return nil, errors.New("plugin_id格式不合法")
	}
	serverRoot, adminRoot, db, closeDB, err := openRemoveResources(options)
	if err != nil {
		return nil, err
	}
	defer closeDB()

	result := &RemoveResult{PluginID: pluginID}
	if err := withPluginDatabaseLock(db, pluginID, func(connection *gorm.DB) error {
		menusRemoved, err := removePluginMetadata(connection, pluginID)
		result.MenusRemoved = menusRemoved
		return err
	}); err != nil {
		return nil, err
	}

	serverTarget := filepath.Join(serverRoot, "internal", "plugins", pluginID)
	adminTarget := filepath.Join(adminRoot, "src", "plugins", pluginID)
	rollback, finalize, removed, err := stageDirectoryRemovals([]string{serverTarget, adminTarget})
	if err != nil {
		return nil, err
	}
	if err := GenerateRegistry(serverRoot); err != nil {
		_ = rollback()
		_ = GenerateRegistry(serverRoot)
		return nil, err
	}
	if err := finalize(); err != nil {
		return nil, fmt.Errorf("清理插件源码备份失败: %w", err)
	}
	result.ServerRemoved = removed[serverTarget]
	result.AdminRemoved = removed[adminTarget]
	return result, nil
}

// Purge 彻底清理插件。先调用Remove使插件不可运行，再删除业务表；只有所有DDL成功后
// 才删除迁移和安装日志。DROP TABLE可重复执行，中断后重新执行可继续完成清理。
func Purge(pluginID string, options RemoveOptions) (*PurgeResult, error) {
	serverRoot, err := filepath.Abs(options.ServerRoot)
	if err != nil {
		return nil, err
	}
	if err := requireArchivedDatabaseDocument(serverRoot, pluginID); err != nil {
		return nil, err
	}
	removed, err := Remove(pluginID, options)
	if err != nil {
		return nil, err
	}
	serverRoot, _, db, closeDB, err := openRemoveResources(options)
	if err != nil {
		return nil, err
	}
	defer closeDB()

	result := &PurgeResult{RemoveResult: *removed}
	err = withPluginDatabaseLock(db, pluginID, func(connection *gorm.DB) error {
		tables, err := pluginBusinessTables(connection, pluginID)
		if err != nil {
			return err
		}
		if err := dropPluginBusinessTables(connection, pluginID, tables); err != nil {
			return err
		}
		result.DroppedTables = append(result.DroppedTables, tables...)
		return connection.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.SysPluginMigration{}).Error; err != nil {
				return err
			}
			return tx.Where("plugin_id = ?", pluginID).Delete(&model.SysPluginInstallLog{}).Error
		})
	})
	if err != nil {
		return nil, err
	}
	if err := os.RemoveAll(filepath.Join(serverRoot, "sql", "plugins", pluginID)); err != nil {
		return nil, fmt.Errorf("删除插件迁移副本失败: %w", err)
	}
	if err := os.RemoveAll(filepath.Join(serverRoot, "plugin_docs", pluginID)); err != nil {
		return nil, fmt.Errorf("删除插件归档文档失败: %w", err)
	}
	return result, nil
}

func requireArchivedDatabaseDocument(serverRoot, pluginID string) error {
	documentRoot := filepath.Join(serverRoot, "plugin_docs", pluginID)
	entries, err := os.ReadDir(documentRoot)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New("缺少已归档的版本数据库说明，拒绝彻底清理；请先人工审查数据库引用")
	}
	if err != nil {
		return fmt.Errorf("读取插件归档文档失败: %w", err)
	}
	latestVersion := ""
	for _, entry := range entries {
		if !entry.IsDir() || !versionPattern.MatchString(entry.Name()) {
			continue
		}
		if latestVersion == "" || compareVersions(entry.Name(), latestVersion) > 0 {
			latestVersion = entry.Name()
		}
	}
	if latestVersion == "" {
		return errors.New("缺少已归档的版本数据库说明，拒绝彻底清理")
	}
	versionRoot := filepath.Join(documentRoot, latestVersion)
	documentPath := filepath.Join(versionRoot, "database", "v"+latestVersion+".md")
	if _, err := os.Stat(documentPath); errors.Is(err, os.ErrNotExist) {
		// 兼容启用版本化文档规则前归档的插件包。
		documentPath = filepath.Join(versionRoot, "DATABASE.md")
	}
	content, err := os.ReadFile(documentPath)
	if err != nil || len(content) == 0 {
		return errors.New("最新归档版本的数据库说明缺失或为空，拒绝彻底清理")
	}
	if err := validateDatabaseDocument(content); err != nil {
		return fmt.Errorf("最新归档版本的数据库说明不完整，拒绝彻底清理: %w", err)
	}
	return nil
}

func openRemoveResources(options RemoveOptions) (string, string, *gorm.DB, func(), error) {
	serverRoot, err := filepath.Abs(options.ServerRoot)
	if err != nil {
		return "", "", nil, func() {}, err
	}
	adminRoot, err := filepath.Abs(options.AdminRoot)
	if err != nil {
		return "", "", nil, func() {}, err
	}
	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		return "", "", nil, func() {}, err
	}
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DSN()), &gorm.Config{})
	if err != nil {
		return "", "", nil, func() {}, fmt.Errorf("连接数据库失败: %w", err)
	}
	closeDB := func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil {
			_ = sqlDB.Close()
		}
	}
	return serverRoot, adminRoot, db, closeDB, nil
}

func removePluginMetadata(db *gorm.DB, pluginID string) (int64, error) {
	for _, table := range []string{"sys_plugin", "sys_plugin_migration", "sys_plugin_menu", "sys_plugin_install_log", "sys_menu", "sys_role_menu"} {
		if !db.Migrator().HasTable(table) {
			return 0, fmt.Errorf("缺少核心表%s，不能安全移除插件", table)
		}
	}
	if err := validateNoEnabledDependents(db, pluginID); err != nil {
		return 0, err
	}
	var pluginRecord model.SysPlugin
	err := db.Where("plugin_id = ?", pluginID).First(&pluginRecord).Error
	if err == nil && pluginRecord.Status != enums.PluginStatusDisabled {
		return 0, errors.New("移除前必须先停用插件并重启服务")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	var menuIDs []uint
	if err := db.Model(&model.SysPluginMenu{}).Where("plugin_id = ?", pluginID).Pluck("menu_id", &menuIDs).Error; err != nil {
		return 0, err
	}
	var menusRemoved int64
	err = db.Transaction(func(tx *gorm.DB) error {
		if len(menuIDs) > 0 {
			if err := tx.Where("menu_id IN ?", menuIDs).Delete(&model.SysRoleMenu{}).Error; err != nil {
				return err
			}
			result := tx.Unscoped().Where("id IN ?", menuIDs).Delete(&model.SysMenu{})
			if result.Error != nil {
				return result.Error
			}
			menusRemoved = result.RowsAffected
		}
		if err := tx.Where("plugin_id = ?", pluginID).Delete(&model.SysPluginMenu{}).Error; err != nil {
			return err
		}
		// 插件信息必须物理删除，否则plugin_id唯一索引会阻止后续重新安装。
		return tx.Unscoped().Where("plugin_id = ?", pluginID).Delete(&model.SysPlugin{}).Error
	})
	return menusRemoved, err
}

func validateNoEnabledDependents(db *gorm.DB, pluginID string) error {
	var records []model.SysPlugin
	if err := db.Where("status = ?", enums.PluginStatusEnabled).Find(&records).Error; err != nil {
		return err
	}
	for _, record := range records {
		if record.PluginID == pluginID {
			continue
		}
		var manifest PackageManifest
		if err := json.Unmarshal([]byte(record.Manifest), &manifest); err != nil {
			return fmt.Errorf("解析已启用插件%s的依赖清单失败", record.PluginID)
		}
		for _, dependency := range manifest.Dependencies {
			if dependency.PluginID == pluginID {
				return fmt.Errorf("已启用插件%s依赖当前插件，不能移除", record.PluginID)
			}
		}
	}
	return nil
}

func pluginBusinessTables(db *gorm.DB, pluginID string) ([]string, error) {
	prefix := "plg_" + pluginID + "_"
	var tables []string
	err := db.Raw("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND LEFT(TABLE_NAME, ?) = ?", len(prefix), prefix).Scan(&tables).Error
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tables)))
	for _, table := range tables {
		if !strings.HasPrefix(table, prefix) {
			return nil, fmt.Errorf("发现不属于插件前缀的表%s", table)
		}
	}
	return tables, nil
}

func dropPluginBusinessTables(db *gorm.DB, pluginID string, tables []string) (returnErr error) {
	if len(tables) == 0 {
		return nil
	}
	prefix := "plg_" + pluginID + "_"
	var externalReferences int64
	err := db.Raw(`SELECT COUNT(*) FROM information_schema.KEY_COLUMN_USAGE
		WHERE REFERENCED_TABLE_SCHEMA = DATABASE()
		  AND LEFT(REFERENCED_TABLE_NAME, ?) = ?
		  AND LEFT(TABLE_NAME, ?) <> ?`, len(prefix), prefix, len(prefix), prefix).Scan(&externalReferences).Error
	if err != nil {
		return err
	}
	if externalReferences > 0 {
		return errors.New("插件业务表仍被插件前缀之外的表通过外键引用，拒绝彻底清理")
	}
	if err := db.Exec("SET SESSION FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		return fmt.Errorf("关闭当前连接外键检查失败: %w", err)
	}
	defer func() {
		if err := db.Exec("SET SESSION FOREIGN_KEY_CHECKS = 1").Error; err != nil && returnErr == nil {
			returnErr = fmt.Errorf("恢复当前连接外键检查失败: %w", err)
		}
	}()
	for _, table := range tables {
		if !strings.HasPrefix(table, prefix) {
			return fmt.Errorf("拒绝删除不属于插件前缀的表%s", table)
		}
		if err := db.Exec("DROP TABLE IF EXISTS `" + table + "`").Error; err != nil {
			return fmt.Errorf("删除插件业务表%s失败: %w", table, err)
		}
	}
	return nil
}

func stageDirectoryRemovals(targets []string) (func() error, func() error, map[string]bool, error) {
	type state struct{ target, backup string }
	states := make([]state, 0, len(targets))
	removed := make(map[string]bool, len(targets))
	rollback := func() error {
		for index := len(states) - 1; index >= 0; index-- {
			if err := os.Rename(states[index].backup, states[index].target); err != nil {
				return err
			}
		}
		return nil
	}
	for _, target := range targets {
		if !pathExists(target) {
			continue
		}
		backup := target + fmt.Sprintf(".plugin-remove-%d", time.Now().UnixNano())
		if err := os.Rename(target, backup); err != nil {
			_ = rollback()
			return nil, nil, nil, err
		}
		states = append(states, state{target: target, backup: backup})
		removed[target] = true
	}
	finalize := func() error {
		for _, item := range states {
			if err := os.RemoveAll(item.backup); err != nil {
				return err
			}
		}
		return nil
	}
	return rollback, finalize, removed, nil
}
