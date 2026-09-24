// Package plugininstaller 提供源码插件发行包校验、安装、升级和静态注册代码生成能力。
package plugininstaller

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	commonplugin "server_api/internal/common/plugin"
)

const maxPackageSize int64 = 100 << 20

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
var versionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// PackageManifest 是插件发行包根目录 plugin.json 的结构。
type PackageManifest struct {
	PluginID     string                    `json:"plugin_id" comment:"插件唯一标识"`
	Name         string                    `json:"name" comment:"插件名称"`
	Version      string                    `json:"version" comment:"插件版本"`
	CoreVersion  string                    `json:"core_version" comment:"兼容核心版本"`
	Logo         string                    `json:"logo" comment:"插件Logo相对路径"`
	Author       string                    `json:"author" comment:"插件作者"`
	Homepage     string                    `json:"homepage" comment:"插件主页"`
	Description  string                    `json:"description" comment:"插件描述"`
	Dependencies []commonplugin.Dependency `json:"dependencies" comment:"插件依赖"`
	Migrations   []PackageMigration        `json:"migrations" comment:"数据库迁移"`
	Menus        []PackageMenu             `json:"menus" comment:"插件菜单"`
}

// PackageMigration 描述发行包内的迁移文件。
type PackageMigration struct {
	Version     string `json:"version" comment:"迁移版本"`
	File        string `json:"file" comment:"迁移SQL相对路径"`
	Checksum    string `json:"checksum" comment:"迁移文件SHA256"`
	Description string `json:"description" comment:"迁移说明"`
}

// PackageMenu 使用稳定业务键描述插件菜单层级，不接收数据库ID。
type PackageMenu struct {
	Key       string `json:"key" comment:"插件内菜单业务键"`
	ParentKey string `json:"parent_key" comment:"插件内父级业务键"`
	Name      string `json:"name" comment:"菜单名称"`
	Type      int    `json:"type" comment:"菜单类型，1目录，2页面，3按钮"`
	Path      string `json:"path" comment:"前端路由"`
	APIPath   string `json:"api_path" comment:"接口权限"`
	Icon      string `json:"icon" comment:"菜单图标"`
	Sort      int    `json:"sort" comment:"排序"`
	Status    int    `json:"status" comment:"显示状态，1显示，2隐藏"`
	Remark    string `json:"remark" comment:"备注"`
}

// Package 是已安全解压并校验的插件发行包。
type Package struct {
	Manifest   PackageManifest
	Root       string
	Hash       string
	cleanupDir string
}

// Close 删除发行包临时解压目录。
func (p *Package) Close() error { return os.RemoveAll(p.cleanupDir) }

// OpenPackage 校验 ZIP 路径、大小、文件摘要和清单后返回临时发行包。
func OpenPackage(archivePath string) (*Package, error) {
	info, err := os.Stat(archivePath)
	if err != nil {
		return nil, fmt.Errorf("读取插件包失败: %w", err)
	}
	if info.Size() > maxPackageSize {
		return nil, errors.New("插件包不能超过100MB")
	}
	hash, err := fileSHA256(archivePath)
	if err != nil {
		return nil, err
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("插件包不是有效ZIP文件: %w", err)
	}
	defer reader.Close()
	if len(reader.File) > 2000 {
		return nil, errors.New("插件包文件数量超过2000")
	}
	var totalUncompressedSize uint64
	for _, file := range reader.File {
		totalUncompressedSize += file.UncompressedSize64
		if totalUncompressedSize > uint64(maxPackageSize) {
			return nil, errors.New("插件包解压后总大小不能超过100MB")
		}
	}
	tempDir, err := os.MkdirTemp("", "server-api-plugin-")
	if err != nil {
		return nil, err
	}
	failed := true
	defer func() {
		if failed {
			_ = os.RemoveAll(tempDir)
		}
	}()
	for _, file := range reader.File {
		cleanName := filepath.Clean(filepath.FromSlash(file.Name))
		if cleanName == "." || filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("插件包包含非法路径 %q", file.Name)
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("插件包不允许包含符号链接 %q", file.Name)
		}
		target := filepath.Join(tempDir, cleanName)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, err
			}
			continue
		}
		if file.UncompressedSize64 > uint64(maxPackageSize) {
			return nil, fmt.Errorf("插件包文件 %s 过大", file.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		input, err := file.Open()
		if err != nil {
			return nil, err
		}
		output, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			_ = input.Close()
			return nil, err
		}
		_, copyErr := io.Copy(output, io.LimitReader(input, maxPackageSize+1))
		closeErr := output.Close()
		_ = input.Close()
		if copyErr != nil || closeErr != nil {
			return nil, errors.New("解压插件包失败")
		}
	}
	manifestData, err := os.ReadFile(filepath.Join(tempDir, "plugin.json"))
	if err != nil {
		return nil, errors.New("插件包根目录缺少plugin.json")
	}
	var manifest PackageManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("解析plugin.json失败: %w", err)
	}
	if err := validatePackage(tempDir, &manifest); err != nil {
		return nil, err
	}
	failed = false
	return &Package{Manifest: manifest, Root: tempDir, Hash: hash, cleanupDir: tempDir}, nil
}

func validatePackage(root string, manifest *PackageManifest) error {
	if !keyPattern.MatchString(manifest.PluginID) {
		return errors.New("plugin_id格式不合法")
	}
	if manifest.Name == "" || manifest.Version == "" {
		return errors.New("插件名称和版本不能为空")
	}
	if !versionPattern.MatchString(manifest.Version) {
		return errors.New("插件版本必须使用语义化版本，例如1.0.0")
	}
	if len(manifest.Logo) > 500 || len(manifest.Author) > 100 || len(manifest.Description) > 2000 {
		return errors.New("插件Logo、作者或描述超过字段长度限制")
	}
	if manifest.Homepage != "" {
		homepage, err := url.Parse(manifest.Homepage)
		if err != nil || (homepage.Scheme != "http" && homepage.Scheme != "https") || homepage.Host == "" || len(manifest.Homepage) > 500 {
			return errors.New("插件主页必须是有效的HTTP或HTTPS地址")
		}
	}
	if manifest.CoreVersion != commonplugin.CoreVersion {
		return fmt.Errorf("插件要求核心版本%s，当前为%s", manifest.CoreVersion, commonplugin.CoreVersion)
	}
	for _, dependency := range manifest.Dependencies {
		if !keyPattern.MatchString(dependency.PluginID) || dependency.PluginID == manifest.PluginID {
			return fmt.Errorf("插件依赖%q不合法", dependency.PluginID)
		}
	}
	serverDir := filepath.Join(root, "server_api", "internal", "plugins", manifest.PluginID)
	if _, err := os.Stat(filepath.Join(serverDir, "plugin.go")); err != nil {
		return errors.New("插件包缺少后端plugin.go")
	}
	seenMenus := make(map[string]struct{}, len(manifest.Menus))
	for _, menu := range manifest.Menus {
		if !keyPattern.MatchString(menu.Key) {
			return fmt.Errorf("菜单业务键%q不合法", menu.Key)
		}
		if _, exists := seenMenus[menu.Key]; exists {
			return fmt.Errorf("菜单业务键%s重复", menu.Key)
		}
		seenMenus[menu.Key] = struct{}{}
		if menu.Type < 1 || menu.Type > 3 || menu.Status < 1 || menu.Status > 2 || menu.Name == "" {
			return fmt.Errorf("菜单%s的类型、状态或名称不合法", menu.Key)
		}
		if menu.Type == 2 && !strings.HasPrefix(menu.Path, "/plugin/"+manifest.PluginID+"/") {
			return fmt.Errorf("菜单%s的页面路径必须位于/plugin/%s/下", menu.Key, manifest.PluginID)
		}
	}
	if _, err := sortMenus(manifest.Menus); err != nil {
		return err
	}
	for _, menu := range manifest.Menus {
		if menu.ParentKey != "" {
			if _, exists := seenMenus[menu.ParentKey]; !exists {
				return fmt.Errorf("菜单%s引用了不存在的父级%s", menu.Key, menu.ParentKey)
			}
		}
	}
	for _, migration := range manifest.Migrations {
		if !versionPattern.MatchString(migration.Version) || migration.File == "" || len(migration.Checksum) != 64 {
			return errors.New("迁移版本、文件或SHA256不合法")
		}
		if !strings.HasPrefix(filepath.ToSlash(filepath.Clean(migration.File)), "database/") {
			return errors.New("迁移文件必须放在发行包database目录")
		}
		migrationPath, err := safeJoin(root, migration.File)
		if err != nil {
			return err
		}
		actual, err := fileSHA256(migrationPath)
		if err != nil {
			return fmt.Errorf("读取迁移%s失败: %w", migration.Version, err)
		}
		if !strings.EqualFold(actual, migration.Checksum) {
			return fmt.Errorf("迁移%s的SHA256不一致", migration.Version)
		}
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return err
		}
		if err := ValidateMigrationSQL(manifest.PluginID, string(content)); err != nil {
			return fmt.Errorf("迁移%s不安全: %w", migration.Version, err)
		}
	}
	return nil
}

func safeJoin(root, name string) (string, error) {
	cleanName := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("非法相对路径%q", name)
	}
	return filepath.Join(root, cleanName), nil
}

func fileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
