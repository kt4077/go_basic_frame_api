package oss

import (
	"context"
	"fmt"
	"io"

	"github.com/qiniu/go-sdk/v7/auth/qbox"
	"github.com/qiniu/go-sdk/v7/storage"
)

// qiniu 七牛云Kodo存储
type qiniu struct {
	bucket   string
	domain   string
	mac      *qbox.Mac
	uploader *storage.FormUploader
}

func newQiniu(params map[string]string) (*qiniu, error) {
	bucket := params["bucket"]
	accessKey := params["access_key"]
	secretKey := params["secret_key"]
	if bucket == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("七牛云配置不完整：需要 bucket/access_key/secret_key")
	}
	mac := qbox.NewMac(accessKey, secretKey)
	cfg := &storage.Config{UseHTTPS: true} // Region 留空由 SDK 按空间自动识别
	domain := params["domain"]
	if domain == "" {
		return nil, fmt.Errorf("七牛云配置不完整：需要 domain（空间绑定的访问域名）")
	}
	return &qiniu{bucket: bucket, domain: domain, mac: mac, uploader: storage.NewFormUploader(cfg)}, nil
}

func (q *qiniu) Upload(ctx context.Context, reader io.Reader, fileName, contentType string) (*UploadResult, error) {
	prepared, err := prepare(reader, fileName, contentType)
	if err != nil {
		return nil, err
	}
	defer prepared.close()
	key := objectKey(prepared.hashName)

	putPolicy := storage.PutPolicy{Scope: fmt.Sprintf("%s:%s", q.bucket, key)}
	token := putPolicy.UploadToken(q.mac)

	ret := storage.PutRet{}
	if err := q.uploader.Put(ctx, &ret, token, key, prepared.file, prepared.size, &storage.PutExtra{
		MimeType: prepared.mimeType,
	}); err != nil {
		return nil, fmt.Errorf("上传七牛云失败: %w", err)
	}
	return &UploadResult{
		FileName:     fileName,
		HashName:     prepared.hashName,
		Size:         prepared.size,
		MimeType:     prepared.mimeType,
		RelativePath: key,
	}, nil
}
