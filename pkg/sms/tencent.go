package sms

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// tencentEndpoint 腾讯云短信默认服务地址。
const tencentEndpoint = "https://sms.tencentcloudapi.com"

// tencentRegion 腾讯云短信默认地域。
const tencentRegion = "ap-guangzhou"

// tencentVersion 腾讯云短信 API 版本。
const tencentVersion = "2021-01-11"

// sendTencent 通过腾讯云短信 v3 接口发送短信，使用官方 TC3-HMAC-SHA256 签名。
func sendTencent(ctx context.Context, cfg Config, msg Message) (*Result, error) {
	phone := "+86" + msg.Phone
	payload := map[string]interface{}{
		"PhoneNumberSet":   []string{phone},
		"SmsSdkAppId":      cfg.AccessKeyID,
		"SignName":         msg.SignName,
		"TemplateId":       msg.TemplateCode,
		"TemplateParamSet": msg.Params,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("构造短信请求失败: %w", err)
	}

	now := time.Now()
	timestamp := now.Unix()
	date := now.UTC().Format("2006-01-02")
	endpoint, err := resolveEndpoint(cfg.Endpoint, tencentEndpoint)
	if err != nil {
		return nil, err
	}
	parsedEndpoint, _ := url.Parse(endpoint)

	authorization, err := tencentAuthorization(cfg.AccessKeySecret, date, string(body), timestamp, parsedEndpoint.Host)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("构造短信请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Host", parsedEndpoint.Host)
	req.Header.Set("X-TC-Action", "SendSms")
	req.Header.Set("X-TC-Version", tencentVersion)
	req.Header.Set("X-TC-Region", tencentRegion)
	req.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", timestamp))
	req.Header.Set("X-TC-Language", "zh-CN")
	req.Header.Set("Authorization", authorization)

	respBody, err := doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	var res struct {
		Response struct {
			SendStatusSet []struct {
				Code     string `json:"Code"`
				Message  string `json:"Message"`
				SerialNo string `json:"SerialNo"`
			} `json:"SendStatusSet"`
			Error *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("解析短信响应失败: %w", err)
	}
	if res.Response.Error != nil {
		return nil, fmt.Errorf("短信发送失败: %s %s", res.Response.Error.Code, res.Response.Error.Message)
	}
	if len(res.Response.SendStatusSet) == 0 {
		return nil, errors.New("短信发送失败: 响应为空")
	}
	if status := res.Response.SendStatusSet[0]; status.Code != "Ok" {
		return nil, fmt.Errorf("短信发送失败: %s %s", status.Code, status.Message)
	}
	return &Result{ProviderMessageID: res.Response.SendStatusSet[0].SerialNo}, nil
}

// tencentAuthorization 计算 TC3-HMAC-SHA256 签名。
func tencentAuthorization(secret, date, body string, timestamp int64, host string) (string, error) {
	hash := func(data []byte) []byte {
		sum := sha256.Sum256(data)
		return sum[:]
	}
	hmacSha256 := func(key, data []byte) []byte {
		mac := hmac.New(sha256.New, key)
		mac.Write(data)
		return mac.Sum(nil)
	}

	canonicalRequest := "POST\n/\n\ncontent-type:application/json; charset=utf-8\nhost:" +
		host + "\n\ncontent-type;host\n" + hex.EncodeToString(hash([]byte(body)))
	stringToSign := "TC3-HMAC-SHA256\n" +
		fmt.Sprintf("%d\n", timestamp) +
		date + "/" + tencentRegion + "/sms/tc3_request\n" +
		hex.EncodeToString(hash([]byte(canonicalRequest)))

	secretDate := hmacSha256([]byte("TC3"+secret), []byte(date))
	secretService := hmacSha256(secretDate, []byte(tencentRegion))
	secretSigning := hmacSha256(secretService, []byte("sms"))
	signature := hex.EncodeToString(hmacSha256(secretSigning, []byte(stringToSign)))
	return fmt.Sprintf("TC3-HMAC-SHA256 Credential=%s/%s/%s/sms, SignedHeaders=content-type;host, Signature=%s",
		secret, date, tencentRegion, signature), nil
}
