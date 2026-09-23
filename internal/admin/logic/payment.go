package logic

import (
	"errors"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/model"
)

type PaymentLogic struct{ App *app.App }

func (l *PaymentLogic) List(c *gin.Context, channel int) ([]resp.PaymentConfigItem, error) {
	var list []model.SysPaymentConfig
	db := l.App.DB.Order("sort ASC, id ASC")
	if channel != 0 {
		db = db.Where("channel = ?", channel)
	}
	if err := db.Find(&list).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return resp.NewPaymentConfigItems(list), nil
}

func (l *PaymentLogic) Save(c *gin.Context, req *param.PaymentConfigSaveReq) (*resp.PaymentConfigItem, error) {
	if req.ID == 0 && req.PrivateKey == "" {
		return nil, errors.New("新增配置必须填写私钥")
	}
	if req.ID == 0 {
		item := model.SysPaymentConfig{Name: req.Name, Channel: req.Channel, AppID: req.AppID, MerchantID: req.MerchantID, PrivateKey: req.PrivateKey, PublicKey: req.PublicKey, APIv3Key: req.APIv3Key, CertSerialNo: req.CertSerialNo, NotifyURL: req.NotifyURL, Status: req.Status, Sort: req.Sort, Remark: req.Remark}
		if err := l.App.DB.Create(&item).Error; err != nil {
			return nil, errors.New("创建失败")
		}
		result := resp.NewPaymentConfigItem(item)
		return &result, nil
	}
	var item model.SysPaymentConfig
	if err := l.App.DB.First(&item, req.ID).Error; err != nil {
		return nil, errors.New("支付配置不存在")
	}
	updates := map[string]interface{}{"name": req.Name, "channel": req.Channel, "app_id": req.AppID, "merchant_id": req.MerchantID, "cert_serial_no": req.CertSerialNo, "notify_url": req.NotifyURL, "status": req.Status, "sort": req.Sort, "remark": req.Remark}
	if req.PrivateKey != "" {
		updates["private_key"] = req.PrivateKey
	}
	if req.PublicKey != "" {
		updates["public_key"] = req.PublicKey
	}
	if req.APIv3Key != "" {
		updates["api_v3_key"] = req.APIv3Key
	}
	if err := l.App.DB.Model(&item).Updates(updates).Error; err != nil {
		return nil, errors.New("修改失败")
	}
	result := resp.NewPaymentConfigItem(item)
	return &result, nil
}

func (l *PaymentLogic) Delete(c *gin.Context, req *param.IDReq) error {
	if result := l.App.DB.Delete(&model.SysPaymentConfig{}, req.ID); result.Error != nil || result.RowsAffected == 0 {
		return errors.New("支付配置不存在或删除失败")
	}
	return nil
}
