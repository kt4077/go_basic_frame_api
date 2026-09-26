package sms

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const smsBaoEndpoint = "https://api.smsbao.com/sms"

// sendSMSBao 通过短信宝国内短信安全接口发送完整短信内容。
// AccessKeyID 对应用户名，AccessKeySecret 对应 API Key 或 MD5 后的平台密码。
func sendSMSBao(ctx context.Context, cfg Config, msg Message) (*Result, error) {
	endpoint, err := resolveEndpoint(cfg.Endpoint, smsBaoEndpoint)
	if err != nil {
		return nil, err
	}
	query := url.Values{
		"u": {cfg.AccessKeyID},
		"p": {cfg.AccessKeySecret},
		"m": {msg.Phone},
		"c": {signedContent(msg.SignName, msg.Content)},
	}
	if msg.TemplateCode != "" {
		query.Set("g", msg.TemplateCode)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("构造短信请求失败: %w", err)
	}
	body, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	code := strings.TrimSpace(string(body))
	if code != "0" {
		return nil, fmt.Errorf("短信发送失败: %s %s", code, smsBaoErrorMessage(code))
	}
	return &Result{}, nil
}

func smsBaoErrorMessage(code string) string {
	messages := map[string]string{
		"30": "密码或 API Key 错误",
		"40": "账号不存在",
		"41": "余额不足",
		"43": "IP 地址受限",
		"50": "内容含有敏感词",
		"51": "手机号码不正确",
	}
	if message := messages[code]; message != "" {
		return message
	}
	return "未知渠道错误"
}
