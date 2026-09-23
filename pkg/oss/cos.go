package oss

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

// tencentCOS 腾讯云COS存储
type tencentCOS struct {
	client *cos.Client
	domain string // 绑定的访问域名（可选，空则用默认Bucket域名）
}

func newCOS(params map[string]string) (*tencentCOS, error) {
	bucket := params["bucket"]
	region := params["region"]
	secretID := params["secret_id"]
	secretKey := params["secret_key"]
	if bucket == "" || region == "" || secretID == "" || secretKey == "" {
		return nil, fmt.Errorf("腾讯云COS配置不完整：需要 region/bucket/secret_id/secret_key")
	}
	bucketURL, err := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", bucket, region))
	if err != nil {
		return nil, fmt.Errorf("解析COS地址失败: %w", err)
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{SecretID: secretID, SecretKey: secretKey},
	})
	domain := params["domain"]
	if domain == "" {
		domain = bucketURL.String()
	}
	return &tencentCOS{client: client, domain: domain}, nil
}

func (c *tencentCOS) Upload(ctx context.Context, reader io.Reader, fileName, contentType string) (*UploadResult, error) {
	prepared, err := prepare(reader, fileName, contentType)
	if err != nil {
		return nil, err
	}
	defer prepared.close()
	key := objectKey(prepared.hashName)
	if _, err := c.client.Object.Put(ctx, key, prepared.file, &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: prepared.mimeType},
	}); err != nil {
		return nil, fmt.Errorf("上传腾讯云COS失败: %w", err)
	}
	return &UploadResult{
		FileName:     fileName,
		HashName:     prepared.hashName,
		Size:         prepared.size,
		MimeType:     prepared.mimeType,
		RelativePath: key,
	}, nil
}
