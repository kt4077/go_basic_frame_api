package plugininstaller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"server_api/config"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
)

// InstallOptions 插件安装或升级参数。
type InstallOptions struct {
	ServerRoot    string
	AdminRoot     string
	ConfigPath    string
	ApplyDatabase bool
	Upgrade       bool
}

// InstallResult 插件安装结果。
type InstallResult struct {
	PluginID        string
	Version         string
	ServerInstalled bool
	AdminInstalled  bool
	DatabaseApplied bool
}

// Install 安装或升级插件源码，自动生成后端静态注册文件，并可选执行受限数据库迁移。
func Install(pkg *Package, options InstallOptions) (*InstallResult, error) {
	serverRoot, err := filepath.Abs(options.ServerRoot)
	if err != nil {
		return nil, err
	}
	adminRoot, err := filepath.Abs(options.AdminRoot)
	if err != nil {
		return nil, err
	}
	pluginID := pkg.Manifest.PluginID
	serverSource := filepath.Join(pkg.Root, "server_api", "internal", "plugins", pluginID)
	serverTarget := filepath.Join(serverRoot, "internal", "plugins", pluginID)
	adminSource := filepath.Join(pkg.Root, "admin_client", "src", "plugins", pluginID)
	adminTarget := filepath.Join(adminRoot, "src", "plugins", pluginID)
	databaseSource := filepath.Join(pkg.Root, "database")
	databaseTarget := filepath.Join(serverRoot, "sql", "plugins", pluginID, pkg.Manifest.Version)
	if err := validateTarget(serverRoot, serverTarget); err != nil {
		return nil, err
	}
	if err := validateTarget(adminRoot, adminTarget); err != nil {
		return nil, err
	}
	if _, err := os.Stat(serverTarget); err == nil && !options.Upgrade {
		return nil, errors.New("后端插件目录已存在，请使用plugin upgrade")
	}
	if _, err := os.Stat(serverTarget); errors.Is(err, os.ErrNotExist) && options.Upgrade {
		return nil, errors.New("后端插件目录不存在，不能执行升级")
	}
	rollback, finalize, err := replaceDirectories([]directoryReplacement{
		{source: serverSource, target: serverTarget, required: true},
		{source: adminSource, target: adminTarget, required: false},
		{source: databaseSource, target: databaseTarget, required: len(pkg.Manifest.Migrations) > 0},
	})
	if err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			_ = rollback()
			_ = GenerateRegistry(serverRoot)
		}
	}()
	if err := GenerateRegistry(serverRoot); err != nil {
		return nil, err
	}
	databaseApplied := false
	if options.ApplyDatabase {
		if err := applyDatabase(pkg, options); err != nil {
			return nil, err
		}
		databaseApplied = true
	}
	committed = true
	if err := finalize(); err != nil {
		return nil, fmt.Errorf("清理插件旧版本备份失败: %w", err)
	}
	return &InstallResult{
		PluginID: pluginID, Version: pkg.Manifest.Version,
		ServerInstalled: true, AdminInstalled: pathExists(adminSource), DatabaseApplied: databaseApplied,
	}, nil
}

type directoryReplacement struct {
	source   string
	target   string
	required bool
}

func replaceDirectories(items []directoryReplacement) (func() error, func() error, error) {
	type state struct {
		target, backup string
		installed      bool
	}
	states := make([]state, 0, len(items))
	rollback := func() error {
		for index := len(states) - 1; index >= 0; index-- {
			item := states[index]
			if item.installed {
				_ = os.RemoveAll(item.target)
			}
			if item.backup != "" {
				_ = os.Rename(item.backup, item.target)
			}
		}
		return nil
	}
	for _, item := range items {
		if !pathExists(item.source) {
			if item.required {
				_ = rollback()
				return nil, nil, fmt.Errorf("插件包缺少目录%s", item.source)
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(item.target), 0o755); err != nil {
			_ = rollback()
			return nil, nil, err
		}
		stage, err := os.MkdirTemp(filepath.Dir(item.target), ".plugin-stage-")
		if err != nil {
			_ = rollback()
			return nil, nil, err
		}
		if err := copyDirectoryContents(item.source, stage); err != nil {
			_ = os.RemoveAll(stage)
			_ = rollback()
			return nil, nil, err
		}
		current := state{target: item.target}
		if pathExists(item.target) {
			current.backup = item.target + fmt.Sprintf(".plugin-backup-%d", time.Now().UnixNano())
			if err := os.Rename(item.target, current.backup); err != nil {
				_ = os.RemoveAll(stage)
				_ = rollback()
				return nil, nil, err
			}
		}
		if err := os.Rename(stage, item.target); err != nil {
			if current.backup != "" {
				_ = os.Rename(current.backup, item.target)
			}
			_ = rollback()
			return nil, nil, err
		}
		current.installed = true
		states = append(states, current)
	}
	finalize := func() error {
		for _, item := range states {
			if item.backup != "" {
				if err := os.RemoveAll(item.backup); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return rollback, finalize, nil
}

func applyDatabase(pkg *Package, options InstallOptions) (returnErr error) {
	cfg, err := config.Load(options.ConfigPath)
	if err != nil {
		return err
	}
	db, err := gorm.Open(mysql.Open(cfg.Mysql.DSN()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	sqlDB, err := db.DB()
	if err == nil {
		defer sqlDB.Close()
	}
	action := enums.PluginInstallActionInstall
	if options.Upgrade {
		action = enums.PluginInstallActionUpgrade
	}
	logRecord := model.SysPluginInstallLog{
		PluginID: pkg.Manifest.PluginID, Version: pkg.Manifest.Version,
		Action: action, Status: enums.PluginInstallStatusFailed, PackageHash: pkg.Hash,
	}
	defer func() {
		if returnErr != nil {
			logRecord.ErrorMessage = returnErr.Error()
			_ = db.Create(&logRecord).Error
		}
	}()
	if err := validateDatabaseState(db, pkg, options.Upgrade); err != nil {
		return err
	}
	for _, migration := range pkg.Manifest.Migrations {
		var existing model.SysPluginMigration
		err := db.Where("plugin_id = ? AND version = ?", pkg.Manifest.PluginID, migration.Version).First(&existing).Error
		if err == nil {
			if existing.Status != enums.PluginMigrationSuccess || existing.Checksum != migration.Checksum {
				return fmt.Errorf("迁移%s已有记录但状态或摘要不一致", migration.Version)
			}
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		migrationPath, _ := safeJoin(pkg.Root, migration.File)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return err
		}
		statements, err := SplitSQLStatements(string(content))
		if err != nil {
			return err
		}
		startedAt := time.Now()
		for _, statement := range statements {
			if err := db.Exec(statement).Error; err != nil {
				return fmt.Errorf("执行迁移%s失败: %w", migration.Version, err)
			}
		}
		record := model.SysPluginMigration{
			PluginID: pkg.Manifest.PluginID, Version: migration.Version,
			Checksum: migration.Checksum, Status: enums.PluginMigrationSuccess,
			ExecutionMs: time.Since(startedAt).Milliseconds(), ExecutedAt: time.Now(),
		}
		if err := db.Create(&record).Error; err != nil {
			return err
		}
	}
	manifestJSON, _ := json.Marshal(pkg.Manifest)
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := applyMenus(tx, pkg.Manifest.PluginID, pkg.Manifest.Menus, options.Upgrade); err != nil {
			return err
		}
		if options.Upgrade {
			return tx.Model(&model.SysPlugin{}).Where("plugin_id = ?", pkg.Manifest.PluginID).Updates(map[string]interface{}{
				"name": pkg.Manifest.Name, "version": pkg.Manifest.Version, "logo": pkg.Manifest.Logo,
				"author": pkg.Manifest.Author, "homepage": pkg.Manifest.Homepage,
				"description": pkg.Manifest.Description, "manifest": string(manifestJSON),
			}).Error
		}
		return tx.Create(&model.SysPlugin{
			PluginID: pkg.Manifest.PluginID, Name: pkg.Manifest.Name, Version: pkg.Manifest.Version,
			Logo: pkg.Manifest.Logo, Author: pkg.Manifest.Author, Homepage: pkg.Manifest.Homepage,
			Description: pkg.Manifest.Description, Status: enums.PluginStatusDisabled,
			Manifest: string(manifestJSON),
		}).Error
	}); err != nil {
		return err
	}
	logRecord.Status = enums.PluginInstallStatusSuccess
	return db.Create(&logRecord).Error
}

func validateDatabaseState(db *gorm.DB, pkg *Package, upgrade bool) error {
	for _, table := range []string{"sys_plugin", "sys_plugin_migration", "sys_plugin_menu", "sys_plugin_install_log", "sys_menu"} {
		if !db.Migrator().HasTable(table) {
			return fmt.Errorf("缺少核心表%s，请先执行v0.0.2升级SQL", table)
		}
	}
	var record model.SysPlugin
	err := db.Where("plugin_id = ?", pkg.Manifest.PluginID).First(&record).Error
	if !upgrade && err == nil {
		return errors.New("插件已安装，请使用plugin upgrade")
	}
	if upgrade {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("插件尚未安装，不能升级")
		}
		if err != nil {
			return err
		}
		if record.Status != enums.PluginStatusDisabled {
			return errors.New("升级前必须先停用插件并重启服务")
		}
		if record.Version == pkg.Manifest.Version {
			return errors.New("目标版本与当前版本相同")
		}
		if compareVersions(pkg.Manifest.Version, record.Version) <= 0 {
			return fmt.Errorf("升级版本%s必须高于当前版本%s", pkg.Manifest.Version, record.Version)
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	for _, dependency := range pkg.Manifest.Dependencies {
		var dependencyRecord model.SysPlugin
		if err := db.Where("plugin_id = ? AND status = ?", dependency.PluginID, enums.PluginStatusEnabled).
			First(&dependencyRecord).Error; err != nil {
			return fmt.Errorf("缺少已启用依赖%s", dependency.PluginID)
		}
		if dependency.Version != "" && dependencyRecord.Version != dependency.Version {
			return fmt.Errorf("依赖%s要求版本%s，当前为%s", dependency.PluginID, dependency.Version, dependencyRecord.Version)
		}
	}
	return nil
}

func compareVersions(left, right string) int {
	leftParts := strings.Split(strings.TrimPrefix(left, "v"), ".")
	rightParts := strings.Split(strings.TrimPrefix(right, "v"), ".")
	length := len(leftParts)
	if len(rightParts) > length {
		length = len(rightParts)
	}
	for index := 0; index < length; index++ {
		leftValue, rightValue := 0, 0
		if index < len(leftParts) {
			leftValue, _ = strconv.Atoi(strings.SplitN(leftParts[index], "-", 2)[0])
		}
		if index < len(rightParts) {
			rightValue, _ = strconv.Atoi(strings.SplitN(rightParts[index], "-", 2)[0])
		}
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func applyMenus(tx *gorm.DB, pluginID string, menus []PackageMenu, upgrade bool) error {
	ordered, err := sortMenus(menus)
	if err != nil {
		return err
	}
	resolved := make(map[string]uint, len(ordered))
	for _, spec := range ordered {
		parentID := uint(0)
		if spec.ParentKey != "" {
			parentID = resolved[spec.ParentKey]
		}
		var mapping model.SysPluginMenu
		err := tx.Where("plugin_id = ? AND menu_key = ?", pluginID, spec.Key).First(&mapping).Error
		if err == nil {
			if !upgrade {
				return fmt.Errorf("菜单业务键%s已经存在", spec.Key)
			}
			updates := map[string]interface{}{
				"name": spec.Name, "type": spec.Type, "parent_id": parentID, "path": spec.Path,
				"api_path": spec.APIPath, "icon": spec.Icon, "sort": spec.Sort,
				"status": spec.Status, "remark": spec.Remark,
			}
			if err := tx.Model(&model.SysMenu{}).Where("id = ?", mapping.MenuID).Updates(updates).Error; err != nil {
				return err
			}
			if err := tx.Model(&mapping).Update("parent_key", spec.ParentKey).Error; err != nil {
				return err
			}
			resolved[spec.Key] = mapping.MenuID
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		menu := model.SysMenu{
			Name: spec.Name, Type: spec.Type, ParentID: parentID, Path: spec.Path,
			ApiPath: spec.APIPath, Icon: spec.Icon, Sort: spec.Sort,
			Status: spec.Status, Remark: spec.Remark,
		}
		if err := tx.Create(&menu).Error; err != nil {
			return err
		}
		mapping = model.SysPluginMenu{PluginID: pluginID, MenuKey: spec.Key, MenuID: menu.ID, ParentKey: spec.ParentKey}
		if err := tx.Create(&mapping).Error; err != nil {
			return err
		}
		resolved[spec.Key] = menu.ID
	}
	return nil
}

func sortMenus(menus []PackageMenu) ([]PackageMenu, error) {
	byKey := make(map[string]PackageMenu, len(menus))
	for _, menu := range menus {
		byKey[menu.Key] = menu
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	states := make(map[string]int, len(keys))
	ordered := make([]PackageMenu, 0, len(keys))
	var visit func(string) error
	visit = func(key string) error {
		if states[key] == 1 {
			return fmt.Errorf("菜单层级存在循环，涉及%s", key)
		}
		if states[key] == 2 {
			return nil
		}
		states[key] = 1
		item := byKey[key]
		if item.ParentKey != "" {
			if _, exists := byKey[item.ParentKey]; !exists {
				return fmt.Errorf("菜单%s父级不存在", key)
			}
			if err := visit(item.ParentKey); err != nil {
				return err
			}
		}
		states[key] = 2
		ordered = append(ordered, item)
		return nil
	}
	for _, key := range keys {
		if err := visit(key); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func validateTarget(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || filepath.IsAbs(relative) || len(relative) >= 3 && relative[:3] == ".."+string(filepath.Separator) {
		return errors.New("插件目标目录越界")
	}
	return nil
}

func copyDirectoryContents(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("插件源码不允许包含符号链接")
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			_ = input.Close()
			return err
		}
		_, copyErr := io.Copy(output, input)
		inputCloseErr := input.Close()
		closeErr := output.Close()
		if copyErr != nil {
			return copyErr
		}
		if inputCloseErr != nil {
			return inputCloseErr
		}
		return closeErr
	})
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
