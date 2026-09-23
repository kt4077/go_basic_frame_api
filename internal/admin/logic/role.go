package logic

import (
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/admin/param"
	"server_api/internal/admin/permission"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/model"
	"server_api/pkg/dberror"
	"server_api/pkg/pagination"
	"server_api/pkg/tree"
)

type RoleLogic struct{ App *app.App }

// Tree 角色树（多级）。
func (l *RoleLogic) Tree(c *gin.Context) ([]*tree.TreeItem[resp.RoleItem], error) {
	var roles []model.SysRole
	if err := l.App.DB.Order("sort ASC, id ASC").Find(&roles).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return tree.BuildTree(resp.NewRoleItems(roles)), nil
}

// List 角色分页列表。
func (l *RoleLogic) List(c *gin.Context, req *param.RoleListReq) (*resp.RoleListRes, error) {
	var total int64
	var roles []model.SysRole
	db := l.App.DB.Model(&model.SysRole{})
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	page, pageSize := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Order("sort ASC, id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&roles).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return &resp.RoleListRes{List: resp.NewRoleItems(roles), Total: total}, nil
}

// Create 新增角色。
func (l *RoleLogic) Create(c *gin.Context, req *param.RoleSaveReq) (*resp.RoleItem, error) {
	var count int64
	l.App.DB.Model(&model.SysRole{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		return nil, errors.New("角色编码已存在")
	}
	role := model.SysRole{
		Name: req.Name, Code: req.Code, ParentID: req.ParentID,
		Sort: req.Sort, Status: req.Status, Remark: req.Remark,
	}
	if err := l.App.DB.Create(&role).Error; err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("角色编码已存在")
		}
		return nil, errors.New("创建失败")
	}
	_ = permission.ClearAllPermissionCache(l.App)
	result := resp.NewRoleItem(role)
	return &result, nil
}

// Update 修改角色。
func (l *RoleLogic) Update(c *gin.Context, req *param.RoleSaveReq) (*resp.RoleItem, error) {
	var role model.SysRole
	if err := l.App.DB.First(&role, req.ID).Error; err != nil {
		return nil, errors.New("角色不存在")
	}
	// 不能把自己挂到自己的子孙节点上
	if req.ParentID != 0 {
		var all []model.SysRole
		_ = l.App.DB.Find(&all).Error
		for _, d := range tree.CollectSelfAndDescendants(all, role.ID) {
			if d == req.ParentID {
				return nil, errors.New("上级角色不能选择自身或其子级")
			}
		}
	}
	var count int64
	l.App.DB.Model(&model.SysRole{}).Where("code = ? AND id != ?", req.Code, role.ID).Count(&count)
	if count > 0 {
		return nil, errors.New("角色编码已存在")
	}
	updates := map[string]interface{}{
		"name": req.Name, "code": req.Code, "parent_id": req.ParentID,
		"sort": req.Sort, "status": req.Status, "remark": req.Remark,
	}
	if err := l.App.DB.Model(&role).Updates(updates).Error; err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("角色编码已存在")
		}
		return nil, errors.New("修改失败")
	}
	_ = permission.ClearAllPermissionCache(l.App)
	result := resp.NewRoleItem(role)
	return &result, nil
}

// Delete 删除角色。删除前校验：有子级角色、有用户绑定则不允许删除。
func (l *RoleLogic) Delete(c *gin.Context, req *param.IDReq) error {
	var role model.SysRole
	if err := l.App.DB.First(&role, req.ID).Error; err != nil {
		return errors.New("角色不存在")
	}
	var childCount int64
	l.App.DB.Model(&model.SysRole{}).Where("parent_id = ?", role.ID).Count(&childCount)
	if childCount > 0 {
		return errors.New("该角色下存在子角色，不能删除")
	}
	var userCount int64
	l.App.DB.Model(&model.SysUserRole{}).Where("role_id = ?", role.ID).Count(&userCount)
	if userCount > 0 {
		return errors.New("该角色已分配给用户，请先解除用户绑定")
	}
	if err := l.App.DB.Delete(&role).Error; err != nil {
		return errors.New("删除失败")
	}
	l.App.DB.Where("role_id = ?", role.ID).Delete(&model.SysRoleMenu{})
	_ = permission.ClearAllPermissionCache(l.App)
	return nil
}

// MenuIDs 角色已绑定的菜单/按钮 ID。
func (l *RoleLogic) MenuIDs(c *gin.Context, req *param.IDReq) ([]uint, error) {
	var ids []uint
	if err := l.App.DB.Model(&model.SysRoleMenu{}).
		Where("role_id = ?", req.ID).Pluck("menu_id", &ids).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return ids, nil
}

// AssignMenus 给角色分配菜单/按钮权限。
func (l *RoleLogic) AssignMenus(c *gin.Context, req *param.AssignMenusReq) error {
	var role model.SysRole
	if err := l.App.DB.First(&role, req.RoleID).Error; err != nil {
		return errors.New("角色不存在")
	}
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", role.ID).Delete(&model.SysRoleMenu{}).Error; err != nil {
			return err
		}
		for _, mid := range req.MenuIDs {
			if err := tx.Create(&model.SysRoleMenu{RoleID: role.ID, MenuID: mid}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return errors.New("分配失败")
	}
	_ = permission.ClearAllPermissionCache(l.App)
	return nil
}

// UserIDs 角色下绑定的用户 ID 列表。
func (l *RoleLogic) UserIDs(c *gin.Context, req *param.IDReq) ([]uint, error) {
	var ids []uint
	if err := l.App.DB.Model(&model.SysUserRole{}).
		Where("role_id = ?", req.ID).Pluck("user_id", &ids).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return ids, nil
}
