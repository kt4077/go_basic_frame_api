package logic

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/common/app"
	"server_api/internal/plugins/news/admin/param"
	"server_api/internal/plugins/news/admin/resp"
	"server_api/internal/plugins/news/model"
	"server_api/pkg/dberror"
	"server_api/pkg/pagination"
)

// CategoryLogic 新闻分类业务逻辑。
type CategoryLogic struct{ App *app.App }

// List 查询新闻分类分页列表。
func (l *CategoryLogic) List(c *gin.Context, req *param.CategoryListReq) (*resp.CategoryListRes, error) {
	var total int64
	var list []resp.CategoryItem
	db := l.App.DB.WithContext(c.Request.Context()).Model(&model.Category{})
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		db = db.Where("name LIKE ? OR slug LIKE ?", like, like)
	}
	if req.Status != 0 {
		db = db.Where("status = ?", req.Status)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询分类失败")
	}
	page, size := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Order("sort DESC, id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, errors.New("查询分类失败")
	}
	return &resp.CategoryListRes{List: list, Total: total}, nil
}

// Save 新增或修改新闻分类。
func (l *CategoryLogic) Save(c *gin.Context, req *param.CategorySaveReq) (*resp.CategoryItem, error) {
	item := model.Category{Name: strings.TrimSpace(req.Name), Slug: strings.ToLower(strings.TrimSpace(req.Slug)), Sort: req.Sort, Status: req.Status, Remark: strings.TrimSpace(req.Remark)}
	err := l.App.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if req.ID == 0 {
			return tx.Create(&item).Error
		}
		var current model.Category
		if err := tx.First(&current, req.ID).Error; err != nil {
			return err
		}
		return tx.Model(&current).Updates(map[string]interface{}{"name": item.Name, "sort": item.Sort, "status": item.Status, "remark": item.Remark}).Error
	})
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("分类标识已存在")
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("分类不存在")
		}
		return nil, errors.New("保存分类失败")
	}
	id := item.ID
	if req.ID != 0 {
		id = req.ID
	}
	var result resp.CategoryItem
	if err := l.App.DB.WithContext(c.Request.Context()).First(&result, id).Error; err != nil {
		return nil, errors.New("读取分类失败")
	}
	return &result, nil
}

// Delete 删除未被文章使用的新闻分类。
func (l *CategoryLogic) Delete(c *gin.Context, req *param.CategoryIDReq) error {
	return l.App.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Article{}).Where("category_id = ?", req.ID).Count(&count).Error; err != nil {
			return errors.New("检查分类引用失败")
		}
		if count > 0 {
			return errors.New("分类下存在文章，不能删除")
		}
		result := tx.Delete(&model.Category{}, req.ID)
		if result.Error != nil {
			return errors.New("删除分类失败")
		}
		if result.RowsAffected == 0 {
			return errors.New("分类不存在")
		}
		return nil
	})
}
