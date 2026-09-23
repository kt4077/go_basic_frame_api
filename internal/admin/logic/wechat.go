package logic

import (
	"errors"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/model"
	"server_api/pkg/dberror"
)

type WechatLogic struct{ App *app.App }

func (l *WechatLogic) List(c *gin.Context, configType int) ([]resp.WechatConfigItem, error) {
	var list []model.SysWechatConfig
	db := l.App.DB.Order("id ASC")
	if configType != 0 {
		db = db.Where("type = ?", configType)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return resp.NewWechatConfigItems(list), nil
}

func (l *WechatLogic) Save(c *gin.Context, req *param.WechatConfigSaveReq) (*resp.WechatConfigItem, error) {
	if req.ID == 0 && req.AppSecret == "" {
		return nil, errors.New("新增配置必须填写 AppSecret")
	}
	if req.ID == 0 {
		item := model.SysWechatConfig{Name: req.Name, Type: req.Type, AppID: req.AppID, AppSecret: req.AppSecret, Token: req.Token, AESKey: req.AESKey, Status: req.Status, Remark: req.Remark}
		if err := l.App.DB.Create(&item).Error; err != nil {
			if dberror.IsDuplicateKey(err) {
				return nil, errors.New("AppID 已存在")
			}
			return nil, errors.New("创建失败")
		}
		result := resp.NewWechatConfigItem(item)
		return &result, nil
	}
	var item model.SysWechatConfig
	if err := l.App.DB.First(&item, req.ID).Error; err != nil {
		return nil, errors.New("微信配置不存在")
	}
	updates := map[string]interface{}{"name": req.Name, "type": req.Type, "app_id": req.AppID, "status": req.Status, "remark": req.Remark}
	if req.AppSecret != "" {
		updates["app_secret"] = req.AppSecret
	}
	if req.Token != "" {
		updates["token"] = req.Token
	}
	if req.AESKey != "" {
		updates["aes_key"] = req.AESKey
	}
	if err := l.App.DB.Model(&item).Updates(updates).Error; err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("AppID 已存在")
		}
		return nil, errors.New("修改失败")
	}
	result := resp.NewWechatConfigItem(item)
	return &result, nil
}

func (l *WechatLogic) Delete(c *gin.Context, req *param.IDReq) error {
	if result := l.App.DB.Delete(&model.SysWechatConfig{}, req.ID); result.Error != nil || result.RowsAffected == 0 {
		return errors.New("微信配置不存在或删除失败")
	}
	return nil
}
