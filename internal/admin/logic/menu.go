package logic

import (
	"errors"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/permission"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/tree"
)

type MenuLogic struct{ App *app.App }

// List 菜单列表（平铺），支持按名称/类型过滤。
func (l *MenuLogic) List(c *gin.Context, req *param.MenuListReq) ([]resp.MenuItem, error) {
	var menus []model.SysMenu
	db := l.App.DB.Order("sort ASC, id ASC")
	if req.Name != "" {
		db = db.Where("name LIKE ?", "%"+req.Name+"%")
	}
	if req.Type != 0 {
		db = db.Where("type = ?", req.Type)
	}
	if err := db.Find(&menus).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return resp.NewMenuItems(menus), nil
}

// Tree 菜单树。
func (l *MenuLogic) Tree(c *gin.Context) ([]*tree.TreeItem[resp.MenuItem], error) {
	var menus []model.SysMenu
	if err := l.App.DB.Order("sort ASC, id ASC").Find(&menus).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return tree.BuildTree(resp.NewMenuItems(menus)), nil
}

// Create 新增菜单/按钮。按钮必须绑定后端接口地址。
func (l *MenuLogic) Create(c *gin.Context, req *param.MenuSaveReq) (*resp.MenuItem, error) {
	if req.Type == enums.MenuTypeButton && req.ApiPath == "" {
		return nil, errors.New("按钮权限必须绑定后端接口地址")
	}
	menu := model.SysMenu{
		Name: req.Name, Type: req.Type, ParentID: req.ParentID,
		Path: req.Path, ApiPath: req.ApiPath, Icon: req.Icon,
		Sort: req.Sort, Status: req.Status, Remark: req.Remark,
	}
	if err := l.App.DB.Create(&menu).Error; err != nil {
		return nil, errors.New("创建失败")
	}
	_ = permission.ClearAllPermissionCache(l.App)
	result := resp.NewMenuItem(menu)
	return &result, nil
}

// Update 修改菜单/按钮。
func (l *MenuLogic) Update(c *gin.Context, req *param.MenuSaveReq) (*resp.MenuItem, error) {
	var menu model.SysMenu
	if err := l.App.DB.First(&menu, req.ID).Error; err != nil {
		return nil, errors.New("菜单不存在")
	}
	if req.Type == enums.MenuTypeButton && req.ApiPath == "" {
		return nil, errors.New("按钮权限必须绑定后端接口地址")
	}
	// 不能把自己挂到自己的子孙节点上
	if req.ParentID != 0 {
		var all []model.SysMenu
		_ = l.App.DB.Find(&all).Error
		for _, d := range tree.CollectSelfAndDescendants(all, menu.ID) {
			if d == req.ParentID {
				return nil, errors.New("上级菜单不能选择自身或其子级")
			}
		}
	}
	updates := map[string]interface{}{
		"name": req.Name, "type": req.Type, "parent_id": req.ParentID,
		"path": req.Path, "api_path": req.ApiPath, "icon": req.Icon,
		"sort": req.Sort, "status": req.Status, "remark": req.Remark,
	}
	if err := l.App.DB.Model(&menu).Updates(updates).Error; err != nil {
		return nil, errors.New("修改失败")
	}
	_ = permission.ClearAllPermissionCache(l.App)
	result := resp.NewMenuItem(menu)
	return &result, nil
}

// Delete 删除菜单/按钮。删除前校验：有子级、有角色绑定则不允许删除。
func (l *MenuLogic) Delete(c *gin.Context, req *param.IDReq) error {
	var menu model.SysMenu
	if err := l.App.DB.First(&menu, req.ID).Error; err != nil {
		return errors.New("菜单不存在")
	}
	var childCount int64
	l.App.DB.Model(&model.SysMenu{}).Where("parent_id = ?", menu.ID).Count(&childCount)
	if childCount > 0 {
		return errors.New("该菜单下存在子级，不能删除")
	}
	var roleCount int64
	l.App.DB.Model(&model.SysRoleMenu{}).Where("menu_id = ?", menu.ID).Count(&roleCount)
	if roleCount > 0 {
		return errors.New("该菜单已被角色使用，请先解除角色绑定")
	}
	if err := l.App.DB.Delete(&menu).Error; err != nil {
		return errors.New("删除失败")
	}
	_ = permission.ClearAllPermissionCache(l.App)
	return nil
}
