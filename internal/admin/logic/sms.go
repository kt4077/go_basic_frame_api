package logic

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	"server_api/pkg/dberror"
	"server_api/pkg/pagination"
	"server_api/pkg/sms"
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
	if req.IsDefault == enums.SMSDefaultYes && req.Status != enums.StatusEnabled {
		return nil, errors.New("默认短信渠道必须为启用状态")
	}
	if req.ID == 0 {
		item := model.SysSMSConfig{Name: req.Name, Provider: req.Provider, AccessKeyID: req.AccessKeyID, AccessKeySecret: req.AccessKeySecret, Endpoint: req.Endpoint, IsDefault: req.IsDefault, Status: req.Status, Remark: req.Remark}
		err := l.App.DB.Transaction(func(tx *gorm.DB) error {
			var configs []model.SysSMSConfig
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id ASC").Find(&configs).Error; err != nil {
				return err
			}
			if len(configs) == 0 {
				if item.Status != enums.StatusEnabled {
					return errors.New("首个短信渠道必须为启用状态")
				}
				item.IsDefault = enums.SMSDefaultYes
			}
			if item.IsDefault == enums.SMSDefaultYes {
				if err := clearOtherSMSDefaults(tx, 0); err != nil {
					return err
				}
			}
			return tx.Create(&item).Error
		})
		if err != nil {
			if dberror.IsDuplicateKey(err) {
				return nil, errors.New("默认短信渠道已被其他请求更新，请刷新后重试")
			}
			if err.Error() == "首个短信渠道必须为启用状态" {
				return nil, err
			}
			return nil, errors.New("创建失败")
		}
		result := resp.NewSMSConfigItem(item)
		return &result, nil
	}
	var item model.SysSMSConfig
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		var configs []model.SysSMSConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id ASC").Find(&configs).Error; err != nil {
			return err
		}
		found := false
		for _, config := range configs {
			if config.ID == req.ID {
				item, found = config, true
				break
			}
		}
		if !found {
			return errors.New("短信配置不存在")
		}
		if req.Provider != enums.SMSProviderYunpian && req.AccessKeySecret == "" && item.AccessKeySecret == "" {
			return errors.New("当前短信服务商必须填写访问密钥")
		}
		if item.IsDefault == enums.SMSDefaultYes &&
			(req.IsDefault != enums.SMSDefaultYes || req.Status != enums.StatusEnabled) {
			return errors.New("默认短信渠道不能取消默认或停用，请先将其他启用渠道设为默认")
		}
		if req.IsDefault == enums.SMSDefaultYes {
			if err := clearOtherSMSDefaults(tx, item.ID); err != nil {
				return err
			}
		}
		updates := map[string]interface{}{
			"name": req.Name, "provider": req.Provider, "access_key_id": req.AccessKeyID,
			"endpoint": req.Endpoint, "is_default": req.IsDefault, "status": req.Status, "remark": req.Remark,
		}
		if req.Provider == enums.SMSProviderYunpian {
			updates["access_key_secret"] = ""
		} else if req.AccessKeySecret != "" {
			updates["access_key_secret"] = req.AccessKeySecret
		}
		return tx.Model(&item).Updates(updates).Error
	})
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("默认短信渠道已被其他请求更新，请刷新后重试")
		}
		for _, message := range []string{"短信配置不存在", "当前短信服务商必须填写访问密钥", "默认短信渠道不能取消默认或停用，请先将其他启用渠道设为默认"} {
			if err.Error() == message {
				return nil, err
			}
		}
		return nil, errors.New("修改失败")
	}
	if err := l.App.DB.First(&item, req.ID).Error; err != nil {
		return nil, errors.New("修改后读取配置失败")
	}
	result := resp.NewSMSConfigItem(item)
	return &result, nil
}

// TestConfig 使用指定渠道及其启用的签名、验证码模板发送测试短信。
// 管理员显式测试不受默认渠道和启用状态限制，但同一配置、手机号一分钟内只能测试一次。
func (l *SMSLogic) TestConfig(c *gin.Context, req *param.SMSConfigTestReq) (*resp.SMSConfigTestRes, error) {
	ctx := c.Request.Context()
	var config model.SysSMSConfig
	if err := l.App.DB.WithContext(ctx).First(&config, req.ConfigID).Error; err != nil {
		return nil, errors.New("短信配置不存在")
	}

	var signature model.SysSMSSignature
	if err := l.App.DB.WithContext(ctx).
		Where("config_id = ? AND status = ? AND sign_code != ''", config.ID, enums.StatusEnabled).
		Order("id ASC").First(&signature).Error; err != nil {
		return nil, errors.New("该渠道未配置启用的短信签名，请先维护签名后再测试")
	}
	var template model.SysSMSTemplate
	if err := l.App.DB.WithContext(ctx).
		Where("config_id = ? AND status = ? AND type = ?", config.ID, enums.StatusEnabled, enums.SMSTemplateVerifyCode).
		Order("id ASC").First(&template).Error; err != nil {
		return nil, errors.New("该渠道未配置启用的验证码模板，请先维护模板后再测试")
	}

	limitKey := fmt.Sprintf("admin:sms_config_test:%d:%s", config.ID, req.Mobile)
	allowed, err := l.App.Redis.SetNX(ctx, limitKey, 1, time.Minute).Result()
	if err != nil {
		return nil, errors.New("短信测试服务异常，请稍后重试")
	}
	if !allowed {
		return nil, errors.New("测试短信发送过于频繁，同一渠道和手机号每分钟只能测试一次")
	}

	code, err := sms.GenerateVerificationCode()
	if err != nil {
		_ = l.App.Redis.Del(ctx, limitKey).Err()
		return nil, errors.New("测试验证码生成失败，请稍后重试")
	}
	content := sms.FillVerificationCode(template.Content, code)
	sentAt := time.Now()
	result, sendErr := sms.Send(ctx, sms.Config{
		Provider:        config.Provider,
		AccessKeyID:     config.AccessKeyID,
		AccessKeySecret: config.AccessKeySecret,
		Endpoint:        config.Endpoint,
	}, sms.Message{
		Phone: req.Mobile, SignName: signature.SignCode, TemplateCode: template.TemplateCode,
		Params: []string{code}, Content: content,
	})

	status, errorMessage, providerMessageID := enums.SMSSendSuccess, "", ""
	if sendErr != nil {
		status, errorMessage = enums.SMSSendFailed, sendErr.Error()
	}
	if result != nil {
		providerMessageID = result.ProviderMessageID
	}
	_ = l.App.DB.WithContext(ctx).Create(&model.SysSMSSendLog{
		ConfigID: config.ID, SignatureID: signature.ID, TemplateID: template.ID,
		Mobile: req.Mobile, Content: content, Status: status,
		ProviderMessageID: providerMessageID, ErrorMessage: errorMessage, SentAt: &sentAt,
	}).Error
	if sendErr != nil {
		return nil, fmt.Errorf("测试短信发送失败：%w", sendErr)
	}
	return &resp.SMSConfigTestRes{ProviderMessageID: providerMessageID}, nil
}

func (l *SMSLogic) DeleteConfig(c *gin.Context, req *param.IDReq) error {
	err := l.App.DB.Transaction(func(tx *gorm.DB) error {
		var configs []model.SysSMSConfig
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Order("id ASC").Find(&configs).Error; err != nil {
			return err
		}
		var config *model.SysSMSConfig
		for index := range configs {
			if configs[index].ID == req.ID {
				config = &configs[index]
				break
			}
		}
		if config == nil {
			return errors.New("短信配置不存在")
		}
		if config.IsDefault == enums.SMSDefaultYes {
			return errors.New("默认短信渠道不允许删除，请先将其他启用渠道设为默认")
		}
		checks := []struct {
			model   interface{}
			message string
		}{
			{&model.SysSMSSignature{}, "该配置下存在签名，请先删除签名"},
			{&model.SysSMSTemplate{}, "该配置下存在模板，请先删除模板"},
			{&model.SysSMSSendLog{}, "该配置已有发送记录，不能删除"},
		}
		for _, check := range checks {
			var count int64
			if err := tx.Model(check.model).Where("config_id = ?", req.ID).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				return errors.New(check.message)
			}
		}
		if result := tx.Delete(config); result.Error != nil || result.RowsAffected == 0 {
			return errors.New("短信配置不存在或删除失败")
		}
		return nil
	})
	if err == nil {
		return nil
	}
	for _, message := range []string{
		"短信配置不存在", "默认短信渠道不允许删除，请先将其他启用渠道设为默认",
		"该配置下存在签名，请先删除签名", "该配置下存在模板，请先删除模板",
		"该配置已有发送记录，不能删除", "短信配置不存在或删除失败",
	} {
		if err.Error() == message {
			return err
		}
	}
	return errors.New("删除失败")
}

// clearOtherSMSDefaults 保证未删除的短信配置中至多一个默认渠道。
func clearOtherSMSDefaults(tx *gorm.DB, id uint) error {
	query := tx.Model(&model.SysSMSConfig{}).Where("is_default = ?", enums.SMSDefaultYes)
	if id > 0 {
		query = query.Where("id != ?", id)
	}
	return query.Update("is_default", enums.SMSDefaultNo).Error
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
