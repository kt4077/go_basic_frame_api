package oss

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// local 本地磁盘存储
type local struct {
	rootPath string // 存储根目录（相对项目目录或绝对路径）
}

func newLocal(params map[string]string) (*local, error) {
	root := params["root_path"]
	if root == "" {
		root = "./uploads"
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("初始化存储目录失败: %w", err)
	}
	return &local{rootPath: root}, nil
}

func (l *local) Upload(_ context.Context, reader io.Reader, fileName, contentType string) (*UploadResult, error) {
	prepared, err := prepare(reader, fileName, contentType)
	if err != nil {
		return nil, err
	}
	defer prepared.close()

	key := objectKey(prepared.hashName)
	// 数据库保存稳定对象Key；磁盘目录仅由 root_path 决定，不进入业务数据。
	diskKey := strings.TrimPrefix(key, "uploads/")
	dir := filepath.Join(l.rootPath, filepath.Dir(diskKey))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	absPath, err := filepath.Abs(filepath.Join(l.rootPath, diskKey))
	if err != nil {
		return nil, fmt.Errorf("计算绝对路径失败: %w", err)
	}
	dst, err := os.OpenFile(absPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}
	_, copyErr := io.Copy(dst, prepared.file)
	closeErr := dst.Close()
	if copyErr != nil {
		return nil, fmt.Errorf("写入文件失败: %w", copyErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("关闭文件失败: %w", closeErr)
	}

	return &UploadResult{
		FileName:     fileName,
		HashName:     prepared.hashName,
		Size:         prepared.size,
		MimeType:     prepared.mimeType,
		RelativePath: key,
	}, nil
}
