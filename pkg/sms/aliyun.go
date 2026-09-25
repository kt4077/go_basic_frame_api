package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// aliyunEndpoint 阿里云短信默认服务地址。
const aliyunEndpoint = "https://dysmsapi.aliyuncs.com"

// sendAliyun 通过阿里云短信 RPC 接口发送短信。
// 签名方式为官方 RPC 风格 HMAC-SHA1，仅使用标准库实现。
func sendAliyun(ctx context.Context, cfg Config, msg Message) error {
	params := map[string]string{
		"Action":           "SendSms",
		"Version":          "2017-05-25",
		"AccessKeyId":      cfg.AccessKeyID,
		"AccessKeySecret":  cfg.AccessKeySecret,
		"SignName":         msg.SignName,
		"TemplateCode":     msg.TemplateCode,
		"PhoneNumbers":     msg.Phone,
		"TemplateParam":    aliyunTemplateParam(msg.Params),
		"RegionId":         "cn-hangzhou",
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureVersion": "1.0",
		"SignatureNonce":   fmt.Sprintf("%d", time.Now().UnixNano()),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Format":           "JSON",
	}

	signParams := aliyunSignParams(params)
	params["Signature"] = aliyunSignature("GET&%2F&"+percentEncode(signParams), cfg.AccessKeySecret+"&")

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = aliyunEndpoint
	}
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = "https://" + endpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+signParams+"&Signature="+url.QueryEscape(params["Signature"]), nil)
	if err != nil {
		return fmt.Errorf("构造短信请求失败: %w", err)
	}
	body, err := doRequest(ctx, req)
	if err != nil {
		return err
	}

	var res struct {
		Code      string `json:"Code"`
		Message   string `json:"Message"`
		RequestID string `json:"RequestId"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("解析短信响应失败: %w", err)
	}
	if res.Code != "OK" {
		return fmt.Errorf("短信发送失败: %s %s", res.Code, res.Message)
	}
	return nil
}

// aliyunTemplateParam 模板变量序列化为 {"code":"123456"} 形式，按 index 命名。
func aliyunTemplateParam(params []string) string {
	entry := make(map[string]string, len(params))
	for i, p := range params {
		entry[fmt.Sprintf("code%d", i+1)] = p
	}
	if len(params) == 1 {
		entry = map[string]string{"code": params[0]}
	}
	data, _ := json.Marshal(entry)
	return string(data)
}

// aliyunSignParams 构造排序并编码后的查询串。
func aliyunSignParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "Signature" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, percentEncode(k)+"="+percentEncode(params[k]))
	}
	return strings.Join(parts, "&")
}

// aliyunSignature 计算阿里云 RPC 签名。
func aliyunSignature(stringToSign, key string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// percentEncode 阿里云 RPC 编码规则：空格转 %20，加号转 %2A，星号转 %2A。
func percentEncode(value string) string {
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	return encoded
}

// doRequest 执行 HTTP 请求并返回响应体。
func doRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("短信服务请求失败")
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, errors.New("读取短信响应失败")
	}
	return body, nil
}
