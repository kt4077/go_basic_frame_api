// Package oss 多平台文件存储：根据存储配置（sys_storage_config）返回对应平台的上传实现。
// 各平台实现统一 Uploader 接口，业务代码只依赖接口，不感知具体平台。
package oss

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"server_api/internal/common/enums"
)

// MaxObjectSize 是存储层的最终文件大小限制。即使调用方未限制 reader，
// 存储适配器也不会无界读取，避免服务层调用导致内存或磁盘耗尽。
const MaxObjectSize int64 = 50 << 20

// UploadResult 上传结果
type UploadResult struct {
	FileName     string // 原始文件名
	HashName     string // 平台hash名（md5 + 扩展名，兼容现有对象命名规则）
	Size         int64  // 文件大小（字节）
	MimeType     string // 文件类型（MIME）
	RelativePath string // 相对路径/对象Key，不包含磁盘根目录或访问域名
}

// Uploader 存储平台上传接口
type Uploader interface {
	Upload(ctx context.Context, reader io.Reader, fileName, contentType string) (*UploadResult, error)
}

// NewUploader 根据渠道类型与参数构造上传实现
func NewUploader(channel string, params map[string]string) (Uploader, error) {
	switch channel {
	case enums.StorageChannelLocal:
		return newLocal(params)
	case enums.StorageChannelAliOSS:
		return newAliOSS(params)
	case enums.StorageChannelCOS:
		return newCOS(params)
	case enums.StorageChannelQiniu:
		return newQiniu(params)
	case enums.StorageChannelMinio:
		return newMinio(params)
	default:
		return nil, fmt.Errorf("不支持的存储渠道: %s", channel)
	}
}

type preparedFile struct {
	file     *os.File
	size     int64
	hashName string
	mimeType string
}

func (p *preparedFile) close() {
	name := p.file.Name()
	_ = p.file.Close()
	_ = os.Remove(name)
}

// prepare 将输入流分块写入临时文件并同步计算摘要，避免把整个文件载入内存。
// 临时文件可重复 seek，兼容各云存储 SDK 对长度和重读的要求。
func prepare(reader io.Reader, fileName, contentType string) (*preparedFile, error) {
	tmp, err := os.CreateTemp("", "go-backend-upload-*")
	if err != nil {
		return nil, fmt.Errorf("创建上传临时文件失败: %w", err)
	}
	failed := true
	defer func() {
		if failed {
			name := tmp.Name()
			_ = tmp.Close()
			_ = os.Remove(name)
		}
	}()

	hash := md5.New()
	written, err := io.Copy(io.MultiWriter(tmp, hash), io.LimitReader(reader, MaxObjectSize+1))
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}
	if written > MaxObjectSize {
		return nil, fmt.Errorf("文件大小超过限制（最大50MB）")
	}

	ext := strings.ToLower(path.Ext(path.Base(fileName)))
	hashName := hex.EncodeToString(hash.Sum(nil)) + ext
	if _, err := tmp.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("读取上传临时文件失败: %w", err)
	}

	mimeType := strings.TrimSpace(strings.Split(contentType, ";")[0])
	if mimeType == "" {
		mimeType = mime.TypeByExtension(ext)
	}
	if mimeType == "" {
		head := make([]byte, 512)
		n, readErr := tmp.Read(head)
		if readErr != nil && readErr != io.EOF {
			return nil, fmt.Errorf("识别文件类型失败: %w", readErr)
		}
		mimeType = http.DetectContentType(head[:n])
		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("读取上传临时文件失败: %w", err)
		}
	}
	// 去掉 MIME 参数，如 text/plain; charset=utf-8，便于存储与后续比较。
	mimeType = strings.TrimSpace(strings.Split(mimeType, ";")[0])

	failed = false
	return &preparedFile{file: tmp, size: written, hashName: hashName, mimeType: mimeType}, nil
}

// objectKey 云端对象Key：uploads/2006/01/02/<hash名>
func objectKey(hashName string) string {
	return "uploads/" + time.Now().Format("2006/01/02") + "/" + hashName
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// httpDetectContentType 内容嗅探（mime 库在部分平台不可用时的兜底）
func httpDetectContentType(data []byte) string {
	if len(data) >= 8 && string(data[:8]) == "\x89PNG\r\n\x1a\n" {
		return "image/png"
	}
	if len(data) >= 3 && string(data[:3]) == "\xff\xd8\xff" {
		return "image/jpeg"
	}
	if len(data) >= 6 && (string(data[:6]) == "GIF87a" || string(data[:6]) == "GIF89a") {
		return "image/gif"
	}
	return "application/octet-stream"
}
