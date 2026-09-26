package logic

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/common/app"
	commonupload "server_api/internal/common/upload"
	"server_api/internal/plugins/news/api/param"
	"server_api/internal/plugins/news/api/resp"
	newsEnums "server_api/internal/plugins/news/enums"
	"server_api/internal/plugins/news/model"
	"server_api/pkg/pagination"
)

// NewsLogic 用户端新闻业务逻辑。
type NewsLogic struct{ App *app.App }

// Categories 查询启用的新闻分类。
func (l *NewsLogic) Categories(c *gin.Context) ([]resp.CategoryItem, error) {
	var list []resp.CategoryItem
	if err := l.App.DB.WithContext(c.Request.Context()).Model(&model.Category{}).Where("status = ?", newsEnums.CategoryStatusEnabled).Order("sort DESC,id ASC").Find(&list).Error; err != nil {
		return nil, errors.New("查询分类失败")
	}
	return list, nil
}

// Articles 查询已发布新闻，分类条件使用稳定标识。
func (l *NewsLogic) Articles(c *gin.Context, req *param.ArticleListReq) (*resp.ArticleListRes, error) {
	var total int64
	var list []resp.ArticleItem
	db := l.App.DB.WithContext(c.Request.Context()).Model(&model.Article{}).
		Where("plg_news_article.status = ? AND plg_news_article.enable_status = ?", newsEnums.ArticleStatusPublished, newsEnums.ArticleEnableStatusEnabled)
	if req.CategorySlug != "" {
		db = db.Joins("JOIN plg_news_category ON plg_news_category.id = plg_news_article.category_id AND plg_news_category.deleted_at IS NULL").Where("plg_news_category.slug = ? AND plg_news_category.status = ?", req.CategorySlug, newsEnums.CategoryStatusEnabled)
	}
	if keyword := strings.TrimSpace(req.Keyword); keyword != "" {
		db = db.Where("plg_news_article.title LIKE ?", "%"+keyword+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询新闻失败")
	}
	page, size := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Preload("Category").Select("plg_news_article.*").Omit("content").Order("sort DESC,published_at DESC,id DESC").Offset((page - 1) * size).Limit(size).Find(&list).Error; err != nil {
		return nil, errors.New("查询新闻失败")
	}
	if err := completeCoverURLs(l.App, list); err != nil {
		return nil, errors.New("封面地址解析失败")
	}
	return &resp.ArticleListRes{List: list, Total: total}, nil
}

// Article 查询已发布文章详情，并以原子更新累加浏览量。
func (l *NewsLogic) Article(c *gin.Context, req *param.ArticleDetailReq) (*resp.ArticleItem, error) {
	var item resp.ArticleItem
	err := l.App.DB.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Category").Where("slug = ? AND status = ? AND enable_status = ?", req.Slug, newsEnums.ArticleStatusPublished, newsEnums.ArticleEnableStatusEnabled).First(&item).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Article{}).Where("id = ?", item.ID).UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
			return err
		}
		item.ViewCount++
		item.TotalViewCount = item.ViewCount + item.VirtualViewCount
		return nil
	})
	if err != nil {
		return nil, errors.New("新闻不存在")
	}
	if err := completeCoverURL(l.App, &item); err != nil {
		return nil, errors.New("封面地址解析失败")
	}
	return &item, nil
}

func completeCoverURLs(application *app.App, list []resp.ArticleItem) error {
	for i := range list {
		if err := completeCoverURL(application, &list[i]); err != nil {
			return err
		}
	}
	return nil
}
func completeCoverURL(application *app.App, item *resp.ArticleItem) error {
	url, err := commonupload.FileURL(application, item.Cover)
	if err != nil {
		return err
	}
	item.CoverURL = url
	item.TotalViewCount = item.ViewCount + item.VirtualViewCount
	return nil
}
