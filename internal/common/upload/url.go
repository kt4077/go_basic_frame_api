package upload

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
)

// URLResolver 根据当前默认存储配置完成相对路径与访问地址之间的转换。
type URLResolver struct {
	domain string
}

// NewURLResolver 创建当前默认存储渠道的文件地址解析器。
func NewURLResolver(application *app.App) (*URLResolver, error) {
	cfg, err := LoadDefaultStorage(application)
	if err != nil {
		return nil, err
	}
	params, err := StorageParams(cfg)
	if err != nil {
		return nil, err
	}
	domain, err := storageDomain(cfg.Channel, params)
	if err != nil {
		return nil, err
	}
	return &URLResolver{domain: strings.TrimRight(domain, "/")}, nil
}

// URL 将数据库中的相对路径补全为可访问地址。
func (r *URLResolver) URL(relativePath string) string {
	if relativePath == "" {
		return ""
	}
	return r.domain + "/" + strings.TrimLeft(relativePath, "/")
}

// Relative 将当前存储域名下的完整地址还原为数据库可保存的相对路径。
func (r *URLResolver) Relative(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if parsed, err := url.ParseRequestURI(value); err == nil && parsed.IsAbs() {
		prefix := r.domain + "/"
		if r.domain == "" || !strings.HasPrefix(value, prefix) {
			return "", errors.New("文件地址不属于当前存储渠道")
		}
		value = strings.TrimPrefix(value, prefix)
	} else if r.domain != "" && strings.HasPrefix(value, r.domain+"/") {
		value = strings.TrimPrefix(value, r.domain+"/")
	}
	value = strings.TrimLeft(strings.ReplaceAll(value, "\\", "/"), "/")
	cleaned := path.Clean(value)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", errors.New("文件相对路径不合法")
	}
	return cleaned, nil
}

// NormalizeFilePath 将上传返回的文件地址规整为可保存的相对路径，
// 供管理端与用户端的业务字段写入前统一调用。
func NormalizeFilePath(application *app.App, value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	resolver, err := NewURLResolver(application)
	if err != nil {
		return "", err
	}
	path, err := resolver.Relative(value)
	if err != nil {
		return "", errors.New("文件地址不合法")
	}
	return path, nil
}

// FileURL 使用当前默认存储配置补全单个文件地址。
func FileURL(application *app.App, relativePath string) (string, error) {
	if relativePath == "" {
		return "", nil
	}
	resolver, err := NewURLResolver(application)
	if err != nil {
		return "", err
	}
	return resolver.URL(relativePath), nil
}

// FileResponse 是上传接口响应，URL 仅动态返回，不写入数据库。
type FileResponse struct {
	ID           uint   `json:"id" comment:"上传记录ID"`
	FileName     string `json:"file_name" comment:"原始文件名"`
	HashName     string `json:"hash_name" comment:"哈希文件名"`
	Size         int64  `json:"size" comment:"文件大小字节"`
	MimeType     string `json:"mime_type" comment:"MIME类型"`
	RelativePath string `json:"relative_path" comment:"文件相对路径"`
	AbsolutePath string `json:"absolute_path" comment:"上传时完整访问地址"`
	URL          string `json:"url" comment:"文件访问地址"`
}

// NewFileResponse 将上传记录转换成带动态访问地址的响应。
func NewFileResponse(application *app.App, record *model.SysUploadFile) (*FileResponse, error) {
	fileURL, err := FileURL(application, record.RelativePath)
	if err != nil {
		return nil, err
	}
	return &FileResponse{ID: record.ID, FileName: record.FileName, HashName: record.HashName, Size: record.Size, MimeType: record.MimeType, RelativePath: record.RelativePath, AbsolutePath: record.AbsolutePath, URL: fileURL}, nil
}

func storageDomain(channel string, params map[string]string) (string, error) {
	if domain := strings.TrimSpace(params["domain"]); domain != "" {
		return domain, nil
	}
	switch channel {
	case enums.StorageChannelLocal:
		return "/files", nil
	case enums.StorageChannelAliOSS:
		endpoint := strings.TrimPrefix(strings.TrimPrefix(params["endpoint"], "https://"), "http://")
		if params["bucket"] == "" || endpoint == "" {
			return "", errors.New("阿里云OSS缺少 bucket 或 endpoint")
		}
		return fmt.Sprintf("https://%s.%s", params["bucket"], endpoint), nil
	case enums.StorageChannelCOS:
		if params["bucket"] == "" || params["region"] == "" {
			return "", errors.New("腾讯云COS缺少 bucket 或 region")
		}
		return fmt.Sprintf("https://%s.cos.%s.myqcloud.com", params["bucket"], params["region"]), nil
	case enums.StorageChannelQiniu:
		return "", errors.New("七牛云存储必须配置访问域名")
	case enums.StorageChannelMinio:
		endpoint := strings.TrimRight(params["endpoint"], "/")
		if endpoint == "" || params["bucket"] == "" {
			return "", errors.New("MinIO缺少 endpoint 或 bucket")
		}
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "http://" + endpoint
		}
		return endpoint + "/" + params["bucket"], nil
	default:
		return "", fmt.Errorf("不支持的存储渠道: %s", channel)
	}
}
