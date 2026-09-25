// 短信验证码业务：发送、限流与校验。验证码为通用能力，与登录注册解耦，
// 由认证、换绑手机号等场景按需调用。
package logic

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server_api/internal/api/param"
	"server_api/internal/common/app"
	"server_api/internal/common/model"
	"server_api/pkg/sms"
)

type SmsLogic struct{ App *app.App }

// 短信验证码场景，与 param 中的 oneof 保持一致。
const (
	smsSceneLogin        = 1 // 登录
	smsSceneRegister     = 2 // 注册
	smsSceneChangeMobile = 3 // 换绑手机号
)

// 验证码安全参数。
const (
	smsCodeTTL     = 5 * time.Minute  // 验证码有效期
	smsMinuteLimit = 2                // 每分钟最多发送条数
	smsWindowLimit = 5                // 5 分钟窗口内最多发送条数
	smsBanTTL      = 30 * time.Minute // 超限后的发送禁令时长
	smsMaxAttempts = 5                // 最大校验错误次数
)

// SendSmsCode 发送短信验证码。限流规则：
// 同一手机号同一场景每分钟最多 2 条；5 分钟内累计超过 5 条则禁止发送 30 分钟。
func (l *SmsLogic) SendSmsCode(c *gin.Context, req *param.SmsCodeReq) error {
	if !mobilePattern.MatchString(req.Mobile) {
		return errors.New("手机号格式不正确")
	}
	ctx := context.Background()
	banKey := fmt.Sprintf("sms_code_ban:%d:%s", req.Scene, req.Mobile)
	if exists, err := l.App.Redis.Exists(ctx, banKey).Result(); err != nil {
		return errors.New("验证码服务异常，请稍后重试")
	} else if exists > 0 {
		l.writeLimitLog(req.Mobile, "触发频控：处于30分钟禁发期")
		return errors.New("验证码发送过于频繁，账号已被限制30分钟，请稍后再试")
	}

	minuteKey := fmt.Sprintf("sms_code_minute:%d:%s:%s", req.Scene, req.Mobile, time.Now().Format("200601021504"))
	count, err := l.App.Redis.Incr(ctx, minuteKey).Result()
	if err != nil {
		return errors.New("验证码服务异常，请稍后重试")
	}
	l.App.Redis.Expire(ctx, minuteKey, 2*time.Minute)
	if count > smsMinuteLimit {
		l.writeLimitLog(req.Mobile, "触发频控：每分钟最多发送2条")
		return errors.New("发送过于频繁，请稍后再试")
	}

	windowKey := fmt.Sprintf("sms_code_window:%d:%s", req.Scene, req.Mobile)
	count, err = l.App.Redis.Incr(ctx, windowKey).Result()
	if err != nil {
		return errors.New("验证码服务异常，请稍后重试")
	}
	if count == 1 {
		l.App.Redis.Expire(ctx, windowKey, smsBanTTL+time.Minute)
	}
	if count > smsWindowLimit {
		l.App.Redis.Set(ctx, banKey, 1, smsBanTTL)
		l.writeLimitLog(req.Mobile, "触发频控：5分钟内发送超过5条，已禁发30分钟")
		return errors.New("验证码发送过于频繁，账号已被限制30分钟，请稍后再试")
	}

	code, err := randomCode()
	if err != nil {
		return errors.New("验证码生成失败，请稍后重试")
	}
	if err := l.sendSms(c, req.Mobile, code); err != nil {
		return err
	}
	l.App.Redis.Set(ctx, fmt.Sprintf("sms_code:%d:%s", req.Scene, req.Mobile), code, smsCodeTTL)
	l.App.Redis.Del(ctx, fmt.Sprintf("sms_code_try:%d:%s", req.Scene, req.Mobile))
	return nil
}

// writeLimitLog 被限流拦截的发送请求同样写入短信发送记录，供管理端审计滥用行为。
// 拦截时尚未读取短信配置，配置/签名/模板ID 记 0。
func (l *SmsLogic) writeLimitLog(mobile, reason string) {
	sentAt := time.Now()
	_ = l.App.DB.Create(&model.SysSMSSendLog{
		Mobile: mobile, Status: 3, ErrorMessage: reason, SentAt: &sentAt,
	}).Error
}

// sendSms 按当前短信配置读取签名与验证码模板并发送，发送结果写入短信发送记录。
func (l *SmsLogic) sendSms(c *gin.Context, mobile, code string) error {
	var config model.SysSMSConfig
	if err := l.App.DB.Where("status = ?", 1).Order("id ASC").First(&config).Error; err != nil {
		return errors.New("短信配置不完整，请联系管理员")
	}
	var signature model.SysSMSSignature
	if err := l.App.DB.Where("config_id = ? AND status = ? AND sign_code != ''", config.ID, 1).
		Order("id ASC").First(&signature).Error; err != nil {
		return errors.New("短信签名未配置，请联系管理员")
	}
	var template model.SysSMSTemplate
	if err := l.App.DB.Where("config_id = ? AND status = ? AND type = ?", config.ID, 1, 1).
		Order("id ASC").First(&template).Error; err != nil {
		return errors.New("短信验证码模板未配置，请联系管理员")
	}

	sentAt := time.Now()
	err := sms.Send(context.Background(), sms.Config{
		Provider:        config.Provider,
		AccessKeyID:     config.AccessKeyID,
		AccessKeySecret: config.AccessKeySecret,
		Endpoint:        config.Endpoint,
	}, sms.Message{
		Phone:        mobile,
		SignName:     signature.SignCode,
		TemplateCode: template.TemplateCode,
		Params:       []string{code},
	})

	status, message := 2, ""
	if err != nil {
		status, message = 3, err.Error()
	}
	sendLog := model.SysSMSSendLog{
		ConfigID: config.ID, SignatureID: signature.ID, TemplateID: template.ID,
		Mobile: mobile, Content: fillTemplateContent(template.Content, code),
		Status: status, ErrorMessage: message, SentAt: &sentAt,
	}
	_ = l.App.DB.Create(&sendLog).Error

	if err != nil {
		return errors.New("短信发送失败，请稍后重试")
	}
	return nil
}

// fillTemplateContent 尽力将验证码填入模板内容，用于发送记录展示。
func fillTemplateContent(content, code string) string {
	for _, placeholder := range []string{"${code}", "{1}", "{code}", "#{code}"} {
		content = strings.ReplaceAll(content, placeholder, code)
	}
	return content
}

// VerifyCode 校验并消费短信验证码；错误次数超限时验证码立即作废。
func (l *SmsLogic) VerifyCode(scene int, mobile, code string) error {
	ctx := context.Background()
	codeKey := fmt.Sprintf("sms_code:%d:%s", scene, mobile)
	stored, err := l.App.Redis.Get(ctx, codeKey).Result()
	if err != nil {
		return errors.New("验证码已失效，请重新获取")
	}
	tryKey := fmt.Sprintf("sms_code_try:%d:%s", scene, mobile)
	attempts, _ := l.App.Redis.Incr(ctx, tryKey).Result()
	l.App.Redis.Expire(ctx, tryKey, smsCodeTTL)
	if attempts > smsMaxAttempts {
		l.App.Redis.Del(ctx, codeKey, tryKey)
		return errors.New("验证码错误次数过多，请重新获取")
	}
	if stored != code {
		return errors.New("验证码错误")
	}
	l.App.Redis.Del(ctx, codeKey, tryKey)
	return nil
}

// randomCode 生成6位数字验证码。
func randomCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
