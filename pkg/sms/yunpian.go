package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const yunpianEndpoint = "https://sms.yunpian.com/v2/sms/single_send.json"

// sendYunpian 通过云片国内短信单条发送接口发送完整短信内容。
// AccessKeyID 对应云片 API Key，AccessKeySecret 不使用。
func sendYunpian(ctx context.Context, cfg Config, msg Message) (*Result, error) {
	endpoint, err := resolveEndpoint(cfg.Endpoint, yunpianEndpoint)
	if err != nil {
		return nil, err
	}
	form := url.Values{
		"apikey": {cfg.AccessKeyID},
		"mobile": {msg.Phone},
		"text":   {signedContent(msg.SignName, msg.Content)},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("构造短信请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	body, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	var response struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		SID  json.Number `json:"sid"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("解析短信响应失败: %w", err)
	}
	if response.Code != 0 {
		return nil, fmt.Errorf("短信发送失败: %d %s", response.Code, response.Msg)
	}
	return &Result{ProviderMessageID: response.SID.String()}, nil
}
