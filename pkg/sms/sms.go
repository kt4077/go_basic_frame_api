// Package sms 短信发送封装，覆盖系统短信配置支持的阿里云、腾讯云渠道。
// 本包只负责按配置发送短信，不包含验证码等业务语义。
package sms

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"server_api/internal/common/enums"
)

// Config 短信渠道配置，来源于 sys_sms_config。
// 渠道枚举统一取自 internal/common/enums，本包不再单独定义。
type Config struct {
	Provider        int    // 渠道：enums.SMSProviderAliyun，enums.SMSProviderTencent
	AccessKeyID     string // 访问密钥ID
	AccessKeySecret string // 访问密钥Secret
	Endpoint        string // 自定义服务地址，留空使用各渠道默认地址
}

// Message 单条短信内容。
type Message struct {
	Phone        string   // 手机号（中国大陆11位）
	SignName     string   // 已审核的短信签名
	TemplateCode string   // 已审核的模板编码
	Params       []string // 模板变量，按模板占位顺序传入
}

// Validate 校验配置与消息的必填项。
func (c Config) Validate() error {
	if c.Provider != enums.SMSProviderAliyun && c.Provider != enums.SMSProviderTencent {
		return errors.New("不支持的短信渠道")
	}
	if c.AccessKeyID == "" || c.AccessKeySecret == "" {
		return errors.New("短信密钥配置不完整")
	}
	return nil
}

// Validate 校验消息必填项。
func (m Message) Validate() error {
	if len(m.Phone) != 11 || !strings.HasPrefix(m.Phone, "1") {
		return errors.New("手机号格式不正确")
	}
	if m.SignName == "" || m.TemplateCode == "" {
		return errors.New("短信签名或模板未配置")
	}
	return nil
}

// Send 按渠道发送短信。发送失败返回携带渠道错误信息的 error。
func Send(ctx context.Context, cfg Config, msg Message) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	if err := msg.Validate(); err != nil {
		return err
	}
	switch cfg.Provider {
	case enums.SMSProviderAliyun:
		return sendAliyun(ctx, cfg, msg)
	case enums.SMSProviderTencent:
		return sendTencent(ctx, cfg, msg)
	default:
		return fmt.Errorf("不支持的短信渠道: %d", cfg.Provider)
	}
}
