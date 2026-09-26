// Package sms 短信发送封装，覆盖系统短信配置支持的全部渠道。
// 本包只负责按配置发送短信，不包含验证码等业务语义。
package sms

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"server_api/internal/common/enums"
)

// Config 短信渠道配置，来源于 sys_sms_config。
// 渠道枚举统一取自 internal/common/enums，本包不再单独定义。
type Config struct {
	Provider        int    // 渠道：取值见 enums.SMSProvider*
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
	Content      string   // 已填充模板变量的完整正文，供内容型短信渠道使用
}

// Result 短信渠道受理结果。
type Result struct {
	ProviderMessageID string // 渠道返回的消息ID，渠道未返回时为空
}

// Validate 校验配置与消息的必填项。
func (c Config) Validate() error {
	switch c.Provider {
	case enums.SMSProviderAliyun, enums.SMSProviderTencent, enums.SMSProviderSMSBao, enums.SMSProviderSMSCN:
		if c.AccessKeyID == "" || c.AccessKeySecret == "" {
			return errors.New("短信密钥配置不完整")
		}
	case enums.SMSProviderYunpian:
		if c.AccessKeyID == "" {
			return errors.New("云片 API Key 未配置")
		}
	default:
		return errors.New("不支持的短信渠道")
	}
	return nil
}

// validateMessage 按渠道校验消息必填项。
func validateMessage(provider int, m Message) error {
	if len(m.Phone) != 11 || !strings.HasPrefix(m.Phone, "1") {
		return errors.New("手机号格式不正确")
	}
	if m.SignName == "" {
		return errors.New("短信签名未配置")
	}
	if provider == enums.SMSProviderAliyun || provider == enums.SMSProviderTencent || provider == enums.SMSProviderSMSCN {
		if m.TemplateCode == "" {
			return errors.New("短信模板未配置")
		}
	}
	if (provider == enums.SMSProviderSMSBao || provider == enums.SMSProviderYunpian) && m.Content == "" {
		return errors.New("短信模板内容未配置")
	}
	return nil
}

// Send 按渠道发送短信。发送失败返回携带渠道错误信息的 error。
func Send(ctx context.Context, cfg Config, msg Message) (*Result, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if err := validateMessage(cfg.Provider, msg); err != nil {
		return nil, err
	}
	switch cfg.Provider {
	case enums.SMSProviderAliyun:
		return sendAliyun(ctx, cfg, msg)
	case enums.SMSProviderTencent:
		return sendTencent(ctx, cfg, msg)
	case enums.SMSProviderSMSBao:
		return sendSMSBao(ctx, cfg, msg)
	case enums.SMSProviderSMSCN:
		return sendSMSCN(ctx, cfg, msg)
	case enums.SMSProviderYunpian:
		return sendYunpian(ctx, cfg, msg)
	default:
		return nil, fmt.Errorf("不支持的短信渠道: %d", cfg.Provider)
	}
}

// signedContent 为需要完整正文的渠道拼接签名，已带括号的签名保持原样。
func signedContent(signName, content string) string {
	signName = strings.TrimSpace(signName)
	signature := signName
	if !strings.HasPrefix(signature, "【") && !strings.HasPrefix(signature, "〖") {
		signature = "【" + signature + "】"
	}
	if strings.HasPrefix(strings.TrimSpace(content), signature) {
		return content
	}
	return signature + content
}

// resolveEndpoint 校验自定义地址，只允许明确的 HTTP/HTTPS 服务地址。
func resolveEndpoint(custom, fallback string) (string, error) {
	endpoint := strings.TrimSpace(custom)
	if endpoint == "" {
		endpoint = fallback
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "https://" + endpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil {
		return "", errors.New("短信服务地址格式不正确")
	}
	return parsed.String(), nil
}
