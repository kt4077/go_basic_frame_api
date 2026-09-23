package permission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/tree"
)

// 权限缓存：key 为用户 ID，value 为该用户可访问的接口集合，多个用换行分隔。
func permCacheKey(userID uint) string { return fmt.Sprintf("perm:%d", userID) }

// RoleIDsOfUser 用户的直接角色 ID 列表。
func RoleIDsOfUser(db *gorm.DB, userID uint) ([]uint, error) {
	var ids []uint
	err := db.Model(&model.SysUserRole{}).Where("user_id = ?", userID).Pluck("role_id", &ids).Error
	return ids, err
}

// ExpandRoleIDs 上级角色自动包含子级、孙级等所有后代角色的 ID。
func ExpandRoleIDs(db *gorm.DB, roleIDs []uint) ([]uint, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var roles []model.SysRole
	if err := db.Find(&roles).Error; err != nil {
		return nil, err
	}
	var out []uint
	for _, id := range roleIDs {
		out = append(out, tree.CollectSelfAndDescendants(roles, id)...)
	}
	return tree.SortUintSlice(out), nil
}

// MenusOfRoles 角色（含后代）绑定的菜单/按钮 ID 列表。
func MenusOfRoles(db *gorm.DB, roleIDs []uint) ([]model.SysMenu, error) {
	if len(roleIDs) == 0 {
		return nil, nil
	}
	var menuIDs []uint
	if err := db.Model(&model.SysRoleMenu{}).
		Where("role_id IN ?", roleIDs).Distinct().Pluck("menu_id", &menuIDs).Error; err != nil {
		return nil, err
	}
	if len(menuIDs) == 0 {
		return nil, nil
	}
	var menus []model.SysMenu
	if err := db.Where("id IN ? AND status = ?", menuIDs, enums.MenuStatusShow).Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

// LoadUserPermissions 计算用户的接口权限集合，如 ["GET:/admin/user/list", ...]。
func LoadUserPermissions(db *gorm.DB, user *model.SysUser) ([]string, error) {
	roleIDs, err := RoleIDsOfUser(db, user.ID)
	if err != nil {
		return nil, err
	}
	allRoleIDs, err := ExpandRoleIDs(db, roleIDs)
	if err != nil {
		return nil, err
	}
	menus, err := MenusOfRoles(db, allRoleIDs)
	if err != nil {
		return nil, err
	}
	var perms []string
	for _, m := range menus {
		for _, api := range strings.Split(m.ApiPath, ",") {
			api = strings.TrimSpace(api)
			if api != "" {
				perms = append(perms, api)
			}
		}
	}
	return perms, nil
}

// CachePermissions 将用户接口权限写入 Redis，权限变更时由管理端主动删除。
func CachePermissions(application *app.App, userID uint, perms []string) error {
	if len(perms) == 0 {
		perms = []string{"__none__"} // 占位，避免缓存穿透
	}
	return application.Redis.Set(context.Background(),
		permCacheKey(userID), strings.Join(perms, "\n"), 24*time.Hour).Err()
}

// ClearPermissionCache 权限/角色/菜单变更后清除相关用户缓存。
func ClearPermissionCache(application *app.App, userIDs ...uint) {
	ctx := context.Background()
	for _, id := range userIDs {
		application.Redis.Del(ctx, permCacheKey(id))
	}
}

// ClearAllPermissionCache 清空所有权限缓存（菜单/角色变更影响面大时使用）。
func ClearAllPermissionCache(application *app.App) error {
	ctx := context.Background()
	iter := application.Redis.Scan(ctx, 0, "perm:*", 100).Iterator()
	for iter.Next(ctx) {
		if err := application.Redis.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// HasPermission 判断用户是否有某个接口权限。超级管理员直接放行。
func HasPermission(ctx context.Context, application *app.App, user *model.SysUser, method, path string) (bool, error) {
	if user.IsSuper == 1 {
		return true, nil
	}
	raw, err := application.Redis.Get(ctx, permCacheKey(user.ID)).Result()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("读取权限缓存失败: %w", err)
	}
	if err == redis.Nil { // 缓存不存在则现算并回填
		perms, perr := LoadUserPermissions(application.DB, user)
		if perr != nil {
			return false, fmt.Errorf("计算用户权限失败: %w", perr)
		}
		if perr := CachePermissions(application, user.ID, perms); perr != nil {
			return false, fmt.Errorf("写入权限缓存失败: %w", perr)
		}
		raw = strings.Join(perms, "\n")
	}
	key := method + ":" + path
	for _, p := range strings.Split(raw, "\n") {
		if p == key {
			return true, nil
		}
	}
	return false, nil
}
