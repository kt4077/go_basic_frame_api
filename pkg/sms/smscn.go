package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const smsCNEndpoint = "https://api.sms.cn/sms/"

// sendSMSCN 通过 SMS.cn JSON 变量模板接口发送短信。
// AccessKeyID 对应 uid，AccessKeySecret 对应平台要求的 32 位 MD5 接口密码。
func sendSMSCN(ctx context.Context, cfg Config, msg Message) (*Result, error) {
	endpoint, err := resolveEndpoint(cfg.Endpoint, smsCNEndpoint)
	if err != nil {
		return nil, err
	}
	params, err := json.Marshal(templateParams(msg.Params))
	if err != nil {
		return nil, fmt.Errorf("构造短信模板变量失败: %w", err)
	}
	form := url.Values{
		"ac":       {"send"},
		"uid":      {cfg.AccessKeyID},
		"pwd":      {cfg.AccessKeySecret},
		"mobile":   {msg.Phone},
		"content":  {string(params)},
		"template": {msg.TemplateCode},
		"format":   {"json"},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("构造短信请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	body, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var response struct {
		Stat    interface{} `json:"stat"`
		Message string      `json:"message"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析短信响应失败: %w", err)
	}
	if fmt.Sprint(response.Stat) != "100" {
		return nil, fmt.Errorf("短信发送失败: %v %s", response.Stat, response.Message)
	}
	return &Result{}, nil
}

// templateParams 将位置参数转换为常见的 code/code2 变量名。
func templateParams(params []string) map[string]string {
	values := make(map[string]string, len(params))
	for index, value := range params {
		key := fmt.Sprintf("code%d", index+1)
		if index == 0 {
			key = "code"
		}
		values[key] = value
	}
	return values
}
