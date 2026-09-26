package logic

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/admin/param"
	"server_api/internal/admin/permission"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	commonupload "server_api/internal/common/upload"
	"server_api/pkg/dberror"
	"server_api/pkg/pagination"
	"server_api/pkg/password"
	"server_api/pkg/tree"
)

type UserLogic struct{ App *app.App }

// List 用户分页列表，支持按关键字、状态、部门（含子部门）过滤。
func (l *UserLogic) List(c *gin.Context, req *param.UserListReq) (*resp.UserListRes, error) {
	var total int64
	var users []resp.UserItem
	db := l.App.DB.Model(&resp.UserItem{})
	if req.Keyword != "" {
		db = db.Where("username LIKE ? OR nickname LIKE ? OR mobile LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.Status != 0 {
		db = db.Where("status = ?", req.Status)
	}
	if req.DeptID != 0 {
		// 选中部门时包含其所有子部门下的人员
		var depts []model.SysDept
		_ = l.App.DB.Find(&depts).Error
		db = db.Where("dept_id IN ?", tree.CollectSelfAndDescendants(depts, req.DeptID))
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	page, pageSize := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC, id ASC")
	}).Order("id ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	if err := completeUserAvatarList(l.App, users); err != nil {
		return nil, errors.New("头像地址解析失败")
	}
	return &resp.UserListRes{List: users, Total: total}, nil
}

// Create 新增用户。
func (l *UserLogic) Create(c *gin.Context, req *param.UserSaveReq) (*resp.UserItem, error) {
	if req.Username == "" || len(req.Password) < 6 {
		return nil, errors.New("参数错误，密码至少6位")
	}
	avatar, err := commonupload.NormalizeFilePath(l.App, req.Avatar)
	if err != nil {
		return nil, err
	}
	var count int64
	l.App.DB.Model(&model.SysUser{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	user := model.SysUser{
		Username: req.Username, Password: hash, Nickname: req.Nickname,
		Avatar: avatar, Mobile: req.Mobile, Email: req.Email, DeptID: req.DeptID,
		Status: req.Status, IsSuper: req.IsSuper,
	}
	err = l.App.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return replaceUserRoles(tx, user.ID, req.RoleIDs)
	})
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("用户名已存在或角色重复")
		}
		return nil, errors.New("创建失败")
	}
	return loadUserItem(l.App, user.ID)
}

// Update 修改用户基本信息与角色。
func (l *UserLogic) Update(c *gin.Context, req *param.UserSaveReq) (*resp.UserItem, error) {
	avatar, err := commonupload.NormalizeFilePath(l.App, req.Avatar)
	if err != nil {
		return nil, err
	}
	var user model.SysUser
	if err := l.App.DB.First(&user, req.ID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	updates := map[string]interface{}{
		"nickname": req.Nickname, "avatar": avatar, "mobile": req.Mobile, "email": req.Email,
		"dept_id": req.DeptID, "status": req.Status, "is_super": req.IsSuper,
	}
	err = l.App.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Updates(updates).Error; err != nil {
			return err
		}
		return replaceUserRoles(tx, user.ID, req.RoleIDs)
	})
	if err != nil {
		return nil, errors.New("修改失败")
	}
	// 权限可能变化，清缓存让在线会话按新权限生效；禁用后立即踢下线
	permission.ClearPermissionCache(l.App, user.ID)
	if req.Status != enums.StatusEnabled {
		_ = auth.KickUser(l.App, user.ID)
	}
	return loadUserItem(l.App, user.ID)
}

// ResetPassword 管理员重置用户密码，重置后踢掉该用户所有会话。
func (l *UserLogic) ResetPassword(c *gin.Context, req *param.ResetPasswordReq) error {
	var user model.SysUser
	if err := l.App.DB.First(&user, req.ID).Error; err != nil {
		return errors.New("用户不存在")
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := l.App.DB.Model(&user).Update("password", hash).Error; err != nil {
		return errors.New("重置失败")
	}
	return auth.KickUser(l.App, user.ID)
}

// Kick 踢用户下线：删除其在线会话的 Redis login_id 缓存。
// 超级管理员不允许被踢下线。
func (l *UserLogic) Kick(c *gin.Context, req *param.IDReq) error {
	var user model.SysUser
	if err := l.App.DB.First(&user, req.ID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.IsSuper == enums.IsSuperYes {
		return errors.New("超级管理员不允许被踢下线")
	}
	return auth.KickUser(l.App, req.ID)
}

// Delete 删除用户。超级管理员不允许删除。
func (l *UserLogic) Delete(c *gin.Context, req *param.IDReq) error {
	var user model.SysUser
	if err := l.App.DB.First(&user, req.ID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if user.IsSuper == enums.IsSuperYes {
		return errors.New("超级管理员不允许删除")
	}
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&user).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", user.ID).Delete(&model.SysUserRole{}).Error
	})
	if err != nil {
		return errors.New("删除失败")
	}
	return auth.KickUser(l.App, user.ID)
}

// replaceUserRoles 覆盖式设置用户角色。
func replaceUserRoles(tx *gorm.DB, userID uint, roleIDs []uint) error {
	if err := tx.Where("user_id = ?", userID).Delete(&model.SysUserRole{}).Error; err != nil {
		return err
	}
	for _, rid := range roleIDs {
		if err := tx.Create(&model.SysUserRole{UserID: userID, RoleID: rid}).Error; err != nil {
			return err
		}
	}
	return nil
}

// loadUserItem 使用 resp 中声明的关联关系预加载用户角色。
func loadUserItem(application *app.App, userID uint) (*resp.UserItem, error) {
	var item resp.UserItem
	if err := application.DB.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort ASC, id ASC")
	}).First(&item, userID).Error; err != nil {
		return nil, err
	}
	if err := completeUserAvatar(application, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func completeUserAvatar(application *app.App, item *resp.UserItem) error {
	if item.Avatar == "" {
		return nil
	}
	fileURL, err := commonupload.FileURL(application, item.Avatar)
	if err != nil {
		return err
	}
	item.Avatar = fileURL
	return nil
}

func completeUserAvatarList(application *app.App, list []resp.UserItem) error {
	resolver, err := commonupload.NewURLResolver(application)
	if err != nil {
		for _, item := range list {
			if item.Avatar != "" {
				return err
			}
		}
		return nil
	}
	for i := range list {
		list[i].Avatar = resolver.URL(list[i].Avatar)
	}
	return nil
}
