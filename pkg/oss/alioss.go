package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// aliOSS 阿里云OSS存储
type aliOSS struct {
	bucket *oss.Bucket
	domain string // 绑定的访问域名（可选，空则用 endpoint + bucket 拼接）
}

func newAliOSS(params map[string]string) (*aliOSS, error) {
	endpoint := params["endpoint"]
	bucketName := params["bucket"]
	accessKeyID := params["access_key_id"]
	accessKeySecret := params["access_key_secret"]
	if endpoint == "" || bucketName == "" || accessKeyID == "" || accessKeySecret == "" {
		return nil, fmt.Errorf("阿里云OSS配置不完整：需要 endpoint/bucket/access_key_id/access_key_secret")
	}
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("连接阿里云OSS失败: %w", err)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("获取Bucket失败: %w", err)
	}
	domain := params["domain"]
	if domain == "" {
		domain = fmt.Sprintf("https://%s.%s", bucketName, endpoint)
	}
	return &aliOSS{bucket: bucket, domain: domain}, nil
}

func (a *aliOSS) Upload(_ context.Context, reader io.Reader, fileName, contentType string) (*UploadResult, error) {
	prepared, err := prepare(reader, fileName, contentType)
	if err != nil {
		return nil, err
	}
	defer prepared.close()
	key := objectKey(prepared.hashName)
	if err := a.bucket.PutObject(key, prepared.file, oss.ContentType(prepared.mimeType)); err != nil {
		return nil, fmt.Errorf("上传阿里云OSS失败: %w", err)
	}
	return &UploadResult{
		FileName:     fileName,
		HashName:     prepared.hashName,
		Size:         prepared.size,
		MimeType:     prepared.mimeType,
		RelativePath: key,
	}, nil
}
