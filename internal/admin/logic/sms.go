package logic

import (
	"errors"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/pagination"
)

type SMSLogic struct{ App *app.App }

func (l *SMSLogic) ConfigList(c *gin.Context) ([]resp.SMSConfigItem, error) {
	var list []model.SysSMSConfig
	err := l.App.DB.Order("id ASC").Find(&list).Error
	return resp.NewSMSConfigItems(list), err
}

func (l *SMSLogic) SaveConfig(c *gin.Context, req *param.SMSConfigSaveReq) (*resp.SMSConfigItem, error) {
	if req.ID == 0 && req.Provider != enums.SMSProviderYunpian && req.AccessKeySecret == "" {
		return nil, errors.New("新增配置必须填写访问密钥")
	}
	if req.ID == 0 {
		item := model.SysSMSConfig{Name: req.Name, Provider: req.Provider, AccessKeyID: req.AccessKeyID, AccessKeySecret: req.AccessKeySecret, Endpoint: req.Endpoint, Status: req.Status, Remark: req.Remark}
		if err := l.App.DB.Create(&item).Error; err != nil {
			return nil, errors.New("创建失败")
		}
		result := resp.NewSMSConfigItem(item)
		return &result, nil
	}
	var item model.SysSMSConfig
	if err := l.App.DB.First(&item, req.ID).Error; err != nil {
		return nil, errors.New("短信配置不存在")
	}
	if req.Provider != enums.SMSProviderYunpian && req.AccessKeySecret == "" && item.AccessKeySecret == "" {
		return nil, errors.New("当前短信服务商必须填写访问密钥")
	}
	updates := map[string]interface{}{"name": req.Name, "provider": req.Provider, "access_key_id": req.AccessKeyID, "endpoint": req.Endpoint, "status": req.Status, "remark": req.Remark}
	if req.Provider == enums.SMSProviderYunpian {
		updates["access_key_secret"] = ""
	} else if req.AccessKeySecret != "" {
		updates["access_key_secret"] = req.AccessKeySecret
	}
	if err := l.App.DB.Model(&item).Updates(updates).Error; err != nil {
		return nil, errors.New("修改失败")
	}
	result := resp.NewSMSConfigItem(item)
	return &result, nil
}

func (l *SMSLogic) DeleteConfig(c *gin.Context, req *param.IDReq) error {
	var count int64
	l.App.DB.Model(&model.SysSMSSignature{}).Where("config_id = ?", req.ID).Count(&count)
	if count > 0 {
		return errors.New("该配置下存在签名，请先删除签名")
	}
	l.App.DB.Model(&model.SysSMSTemplate{}).Where("config_id = ?", req.ID).Count(&count)
	if count > 0 {
		return errors.New("该配置下存在模板，请先删除模板")
	}
	l.App.DB.Model(&model.SysSMSSendLog{}).Where("config_id = ?", req.ID).Count(&count)
	if count > 0 {
		return errors.New("该配置已有发送记录，不能删除")
	}
	if result := l.App.DB.Delete(&model.SysSMSConfig{}, req.ID); result.Error != nil || result.RowsAffected == 0 {
		return errors.New("短信配置不存在或删除失败")
	}
	return nil
}

func (l *SMSLogic) SignatureList(c *gin.Context) ([]resp.SMSSignatureItem, error) {
	var list []model.SysSMSSignature
	err := l.App.DB.Order("id ASC").Find(&list).Error
	return resp.NewSMSSignatureItems(list), err
}

func (l *SMSLogic) SaveSignature(c *gin.Context, req *param.SMSSignatureSaveReq) (*resp.SMSSignatureItem, error) {
	var config model.SysSMSConfig
	if err := l.App.DB.First(&config, req.ConfigID).Error; err != nil {
		return nil, errors.New("短信开发配置不存在")
	}
	item := model.SysSMSSignature{ConfigID: req.ConfigID, Name: req.Name, SignCode: req.SignCode, Status: req.Status, Remark: req.Remark}
	if req.ID == 0 {
		if err := l.App.DB.Create(&item).Error; err != nil {
			return nil, errors.New("创建失败")
		}
	} else {
		if result := l.App.DB.Model(&model.SysSMSSignature{}).Where("id = ?", req.ID).
			Select("config_id", "name", "sign_code", "status", "remark").Updates(&item); result.Error != nil || result.RowsAffected == 0 {
			return nil, errors.New("签名不存在或修改失败")
		}
		item.ID = req.ID
	}
	result := resp.NewSMSSignatureItem(item)
	return &result, nil
}

func (l *SMSLogic) DeleteSignature(c *gin.Context, req *param.IDReq) error {
	var count int64
	l.App.DB.Model(&model.SysSMSSendLog{}).Where("signature_id = ?", req.ID).Count(&count)
	if count > 0 {
		return errors.New("该签名已有发送记录，不能删除")
	}
	if result := l.App.DB.Delete(&model.SysSMSSignature{}, req.ID); result.Error != nil || result.RowsAffected == 0 {
		return errors.New("签名不存在或删除失败")
	}
	return nil
}

func (l *SMSLogic) TemplateList(c *gin.Context) ([]resp.SMSTemplateItem, error) {
	var list []model.SysSMSTemplate
	err := l.App.DB.Order("id ASC").Find(&list).Error
	return resp.NewSMSTemplateItems(list), err
}

func (l *SMSLogic) SaveTemplate(c *gin.Context, req *param.SMSTemplateSaveReq) (*resp.SMSTemplateItem, error) {
	var config model.SysSMSConfig
	if err := l.App.DB.First(&config, req.ConfigID).Error; err != nil {
		return nil, errors.New("短信开发配置不存在")
	}
	if req.TemplateCode == "" && config.Provider != enums.SMSProviderSMSBao && config.Provider != enums.SMSProviderYunpian {
		return nil, errors.New("当前短信服务商必须填写平台模板编码")
	}
	item := model.SysSMSTemplate{ConfigID: req.ConfigID, Name: req.Name, TemplateCode: req.TemplateCode, Type: req.Type, Content: req.Content, Status: req.Status, Remark: req.Remark}
	if req.ID == 0 {
		if err := l.App.DB.Create(&item).Error; err != nil {
			return nil, errors.New("创建失败")
		}
	} else {
		if result := l.App.DB.Model(&model.SysSMSTemplate{}).Where("id = ?", req.ID).
			Select("config_id", "name", "template_code", "type", "content", "status", "remark").Updates(&item); result.Error != nil || result.RowsAffected == 0 {
			return nil, errors.New("模板不存在或修改失败")
		}
		item.ID = req.ID
	}
	result := resp.NewSMSTemplateItem(item)
	return &result, nil
}

func (l *SMSLogic) DeleteTemplate(c *gin.Context, req *param.IDReq) error {
	var count int64
	l.App.DB.Model(&model.SysSMSSendLog{}).Where("template_id = ?", req.ID).Count(&count)
	if count > 0 {
		return errors.New("该模板已有发送记录，不能删除")
	}
	if result := l.App.DB.Delete(&model.SysSMSTemplate{}, req.ID); result.Error != nil || result.RowsAffected == 0 {
		return errors.New("模板不存在或删除失败")
	}
	return nil
}

func (l *SMSLogic) LogList(c *gin.Context, req *param.SMSLogListReq) (*resp.SMSLogListRes, error) {
	db := l.App.DB.Model(&model.SysSMSSendLog{})
	if req.Mobile != "" {
		db = db.Where("mobile LIKE ?", "%"+req.Mobile+"%")
	}
	if req.Status != 0 {
		db = db.Where("status = ?", req.Status)
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	page, pageSize := pagination.Normalize(req.Page, req.PageSize)
	var list []model.SysSMSSendLog
	if err := db.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	return &resp.SMSLogListRes{List: resp.NewSMSLogItems(list), Total: total}, nil
}
