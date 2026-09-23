package logic

import (
	"errors"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/model"
	"server_api/pkg/tree"
)

type DeptLogic struct{ App *app.App }

// Tree 部门树（多级）。
func (l *DeptLogic) Tree(c *gin.Context) ([]*tree.TreeItem[resp.DeptItem], error) {
	var depts []model.SysDept
	if err := l.App.DB.Order("sort ASC, id ASC").Find(&depts).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return tree.BuildTree(resp.NewDeptItems(depts)), nil
}

// Create 新增部门。
func (l *DeptLogic) Create(c *gin.Context, req *param.DeptSaveReq) (*resp.DeptItem, error) {
	dept := model.SysDept{
		Name: req.Name, ParentID: req.ParentID,
		Sort: req.Sort, Leader: req.Leader, Remark: req.Remark,
	}
	if err := l.App.DB.Create(&dept).Error; err != nil {
		return nil, errors.New("创建失败")
	}
	result := resp.NewDeptItem(dept)
	return &result, nil
}

// Update 修改部门。
func (l *DeptLogic) Update(c *gin.Context, req *param.DeptSaveReq) (*resp.DeptItem, error) {
	var dept model.SysDept
	if err := l.App.DB.First(&dept, req.ID).Error; err != nil {
		return nil, errors.New("部门不存在")
	}
	if req.ParentID != 0 {
		var all []model.SysDept
		_ = l.App.DB.Find(&all).Error
		for _, d := range tree.CollectSelfAndDescendants(all, dept.ID) {
			if d == req.ParentID {
				return nil, errors.New("上级部门不能选择自身或其子级")
			}
		}
	}
	updates := map[string]interface{}{
		"name": req.Name, "parent_id": req.ParentID,
		"sort": req.Sort, "leader": req.Leader, "remark": req.Remark,
	}
	if err := l.App.DB.Model(&dept).Updates(updates).Error; err != nil {
		return nil, errors.New("修改失败")
	}
	result := resp.NewDeptItem(dept)
	return &result, nil
}

// Delete 删除部门。删除前校验：有子部门、有人员归属则不允许删除。
func (l *DeptLogic) Delete(c *gin.Context, req *param.IDReq) error {
	var dept model.SysDept
	if err := l.App.DB.First(&dept, req.ID).Error; err != nil {
		return errors.New("部门不存在")
	}
	var childCount int64
	l.App.DB.Model(&model.SysDept{}).Where("parent_id = ?", dept.ID).Count(&childCount)
	if childCount > 0 {
		return errors.New("该部门下存在子部门，不能删除")
	}
	var userCount int64
	l.App.DB.Model(&model.SysUser{}).Where("dept_id = ?", dept.ID).Count(&userCount)
	if userCount > 0 {
		return errors.New("该部门下存在人员，请先转移人员")
	}
	if err := l.App.DB.Delete(&dept).Error; err != nil {
		return errors.New("删除失败")
	}
	return nil
}
