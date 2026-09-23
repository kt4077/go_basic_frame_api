package oss

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// minio MinIO / S3 兼容存储
type minioStore struct {
	client   *minio.Client
	bucket   string
	domain   string // 绑定的访问域名（可选，空则用 endpoint + bucket 拼接）
	endpoint string
	secure   bool
}

func newMinio(params map[string]string) (*minioStore, error) {
	rawEndpoint := params["endpoint"]
	bucket := params["bucket"]
	accessKey := params["access_key"]
	secretKey := params["secret_key"]
	if rawEndpoint == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("MinIO配置不完整：需要 endpoint/bucket/access_key/secret_key")
	}

	secure := strings.HasPrefix(rawEndpoint, "https://")
	endpoint := strings.TrimPrefix(strings.TrimPrefix(rawEndpoint, "https://"), "http://")
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("连接MinIO失败: %w", err)
	}

	domain := params["domain"]
	if domain == "" {
		scheme := "http"
		if secure {
			scheme = "https"
		}
		domain = fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucket)
	}
	return &minioStore{client: client, bucket: bucket, domain: domain, endpoint: endpoint, secure: secure}, nil
}

func (m *minioStore) Upload(ctx context.Context, reader io.Reader, fileName, contentType string) (*UploadResult, error) {
	prepared, err := prepare(reader, fileName, contentType)
	if err != nil {
		return nil, err
	}
	defer prepared.close()
	key := objectKey(prepared.hashName)
	if _, err := m.client.PutObject(ctx, m.bucket, key, prepared.file, prepared.size,
		minio.PutObjectOptions{ContentType: prepared.mimeType}); err != nil {
		return nil, fmt.Errorf("上传MinIO失败: %w", err)
	}
	return &UploadResult{
		FileName:     fileName,
		HashName:     prepared.hashName,
		Size:         prepared.size,
		MimeType:     prepared.mimeType,
		RelativePath: key,
	}, nil
}
