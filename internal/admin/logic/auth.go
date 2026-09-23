// Package logic 管理端业务逻辑层。controller 调用 logic，统一传入 *gin.Context，
// 登录信息、IP、User-Agent 等上下文数据均在 logic 层获取。
package logic

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/admin/param"
	"server_api/internal/admin/permission"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/password"
	"server_api/pkg/tree"
)

type AuthLogic struct{ App *app.App }

// Login 管理端登录：校验账号、生成 login_id 并写入 JWT 与 Redis。
func (l *AuthLogic) Login(c *gin.Context, req *param.LoginReq) (*resp.LoginRes, error) {
	var user model.SysUser
	err := l.App.DB.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		// 区分真实原因，避免把表不存在、数据未导入等问题误报成密码错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, fmt.Errorf("登录异常，请检查数据库: %w", err)
	}
	if !password.Check(user.Password, req.Password) {
		return nil, errors.New("密码错误")
	}
	if user.Status != enums.StatusEnabled {
		return nil, errors.New("账号已被禁用")
	}
	token, err := auth.CreateLogin(l.App, &user, enums.ClientAdmin, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	return &resp.LoginRes{Token: token}, nil
}

// Logout 退出登录：删除 Redis 中的 login_id，实现 JWT 主动过期。
func (l *AuthLogic) Logout(c *gin.Context) error {
	claims := auth.CtxClaims(c)
	return auth.Logout(l.App, claims.LoginID, enums.LoginStatusLogout)
}

// Me 当前登录用户信息。
func (l *AuthLogic) Me(c *gin.Context) (*resp.UserItem, error) {
	claims := auth.CtxClaims(c)
	user, err := auth.GetUserByID(l.App.DB, claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	result, err := loadUserItem(l.App, user.ID)
	if err != nil {
		return nil, errors.New("查询用户信息失败")
	}
	return result, nil
}

// UpdateProfile 修改当前管理员的个人资料，不允许在此变更权限、部门或账号状态。
func (l *AuthLogic) UpdateProfile(c *gin.Context, req *param.ProfileUpdateReq) (*resp.UserItem, error) {
	avatar, err := normalizeFilePath(l.App, req.Avatar)
	if err != nil {
		return nil, err
	}
	claims := auth.CtxClaims(c)
	updates := map[string]interface{}{
		"nickname": req.Nickname,
		"avatar":   avatar,
		"mobile":   req.Mobile,
		"email":    req.Email,
	}
	result := l.App.DB.Model(&model.SysUser{}).Where("id = ?", claims.UserID).Updates(updates)
	if result.Error != nil {
		return nil, errors.New("个人资料保存失败")
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := l.App.DB.Model(&model.SysUser{}).Where("id = ?", claims.UserID).Count(&count).Error; err != nil || count == 0 {
			return nil, errors.New("用户不存在")
		}
	}
	return loadUserItem(l.App, claims.UserID)
}

// UpdateAvatar 单独修改当前管理员头像，不影响尚未保存的其他个人资料。
func (l *AuthLogic) UpdateAvatar(c *gin.Context, req *param.AvatarUpdateReq) (*resp.UserItem, error) {
	avatar, err := normalizeFilePath(l.App, req.Avatar)
	if err != nil {
		return nil, err
	}
	claims := auth.CtxClaims(c)
	result := l.App.DB.Model(&model.SysUser{}).Where("id = ?", claims.UserID).Update("avatar", avatar)
	if result.Error != nil {
		return nil, errors.New("头像保存失败")
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := l.App.DB.Model(&model.SysUser{}).Where("id = ?", claims.UserID).Count(&count).Error; err != nil || count == 0 {
			return nil, errors.New("用户不存在")
		}
	}
	return loadUserItem(l.App, claims.UserID)
}

// GetRouters 返回当前用户的菜单树（目录+菜单，不含按钮），用于渲染左侧导航。
func (l *AuthLogic) GetRouters(c *gin.Context) ([]*tree.TreeItem[resp.MenuItem], error) {
	claims := auth.CtxClaims(c)
	user, err := auth.GetUserByID(l.App.DB, claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}

	db := l.App.DB.Where("type IN ?", []int{enums.MenuTypeDir, enums.MenuTypePage}).
		Where("status = ?", enums.MenuStatusShow).Order("sort ASC, id ASC")
	if user.IsSuper != 1 {
		menus, err := l.menusOfUser(user)
		if err != nil {
			return nil, err
		}
		if len(menus) == 0 {
			return []*tree.TreeItem[resp.MenuItem]{}, nil
		}
		ids := make([]uint, 0, len(menus))
		for _, m := range menus {
			ids = append(ids, m.ID)
		}
		db = db.Where("id IN ?", ids)
	}
	var menus []model.SysMenu
	if err := db.Find(&menus).Error; err != nil {
		return nil, errors.New("查询菜单失败")
	}
	return tree.BuildTree(resp.NewMenuItems(menus)), nil
}

// ChangePassword 修改自己的密码，成功后踢掉该账号所有在线会话。
func (l *AuthLogic) ChangePassword(c *gin.Context, req *param.ChangePasswordReq) error {
	claims := auth.CtxClaims(c)
	user, err := auth.GetUserByID(l.App.DB, claims.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if !password.Check(user.Password, req.OldPassword) {
		return errors.New("原密码错误")
	}
	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := l.App.DB.Model(user).Update("password", hash).Error; err != nil {
		return errors.New("修改失败")
	}
	return auth.KickUser(l.App, user.ID)
}

// Permissions 当前用户的接口权限列表（前端按钮显隐用），超级管理员返回通配。
func (l *AuthLogic) Permissions(c *gin.Context) ([]string, error) {
	claims := auth.CtxClaims(c)
	user, err := auth.GetUserByID(l.App.DB, claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	if user.IsSuper == 1 {
		return []string{"*"}, nil
	}
	perms, err := permission.LoadUserPermissions(l.App.DB, user)
	if err != nil {
		return nil, errors.New("查询权限失败")
	}
	return perms, nil
}

// menusOfUser 普通用户可见的菜单/按钮列表（含后代角色绑定的菜单）。
func (l *AuthLogic) menusOfUser(user *model.SysUser) ([]model.SysMenu, error) {
	roleIDs, err := permission.RoleIDsOfUser(l.App.DB, user.ID)
	if err != nil {
		return nil, err
	}
	allRoleIDs, err := permission.ExpandRoleIDs(l.App.DB, roleIDs)
	if err != nil {
		return nil, err
	}
	return permission.MenusOfRoles(l.App.DB, allRoleIDs)
}
