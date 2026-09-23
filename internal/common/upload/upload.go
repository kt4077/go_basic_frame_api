// 通用文件上传逻辑：
//   - HTTP multipart 上传：管理端(/admin/upload/file)与用户端(/api/upload/file)共用；
//   - 服务层调用：SaveFileStream 保存服务端获得的二进制流（如微信小程序二维码），
//     SaveRemoteFile 抓取远程地址文件（如第三方图片）并转存到存储平台。
package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
	"server_api/internal/common/model"
	"server_api/pkg/oss"
)

// MaxUploadSize 单文件大小上限：50MB，与存储层的兜底限制保持一致。
const MaxUploadSize = oss.MaxObjectSize

// UploadFile 接收 multipart 文件（字段名 file），上传到当前默认存储平台并写入上传记录。
// client 区分来源端口（admin/api），uploaderID 为上传人ID（可空）。
func UploadFile(application *app.App, c *gin.Context, client string, uploaderID *uint) (*model.SysUploadFile, error) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, errors.New("请选择要上传的文件（form字段名：file）")
	}
	if fileHeader.Size > MaxUploadSize {
		return nil, errors.New("文件大小超过限制（最大50MB）")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("读取上传文件失败")
	}
	defer src.Close()

	uploader, err := LoadUploader(application)
	if err != nil {
		return nil, err
	}

	result, err := uploader.Upload(c.Request.Context(), src, fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	return saveUploadRecord(application, client, uploaderID, result)
}

// UploadFileStream 二进制文件流上传：请求体即文件原始内容（非 multipart）。
// 文件名通过 ?filename= 或请求头 X-Filename 传入，缺省为 stream_<时间戳>.bin；
// Content-Type 请求头可选，缺省时自动嗅探。五种存储渠道均支持。
func UploadFileStream(application *app.App, c *gin.Context, client string, uploaderID *uint) (*model.SysUploadFile, error) {
	fileName := c.Query("filename")
	if fileName == "" {
		fileName = c.GetHeader("X-Filename")
	}
	if fileName == "" {
		fileName = fmt.Sprintf("stream_%d.bin", time.Now().UnixMilli())
	}

	// 限制请求体大小，超限时读取报错
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)

	uploader, err := LoadUploader(application)
	if err != nil {
		return nil, err
	}

	result, err := uploader.Upload(c.Request.Context(), c.Request.Body, fileName, c.GetHeader("Content-Type"))
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return nil, errors.New("文件大小超过限制（最大50MB）")
		}
		return nil, err
	}
	return saveUploadRecord(application, client, uploaderID, result)
}

// saveUploadRecord 将上传结果写入记录表
func saveUploadRecord(application *app.App, client string, uploaderID *uint, result *oss.UploadResult) (*model.SysUploadFile, error) {
	if result.MimeType == "" {
		result.MimeType = mime.TypeByExtension(path.Ext(result.FileName))
	}
	resolver, err := NewURLResolver(application)
	if err != nil {
		return nil, err
	}
	record := &model.SysUploadFile{
		FileName:     result.FileName,
		HashName:     result.HashName,
		Size:         result.Size,
		MimeType:     result.MimeType,
		Ext:          strings.TrimPrefix(strings.ToLower(path.Ext(result.FileName)), "."),
		Client:       client,
		RelativePath: result.RelativePath,
		AbsolutePath: resolver.URL(result.RelativePath),
		UploaderID:   uploaderID,
	}
	if err := application.DB.Create(record).Error; err != nil {
		return nil, errors.New("上传记录写入失败")
	}
	return record, nil
}

// SaveFileStream 服务层文件流上传：把服务端获得的二进制流保存到当前默认存储平台，并写入上传记录。
// 典型场景：获取到微信小程序二维码的文件流、程序内生成的文件等。
// fileName 建议带扩展名；contentType 可空（自动嗅探）；uploaderID 可空（系统内部上传时传 nil）。
func SaveFileStream(application *app.App, client, fileName, contentType string, reader io.Reader, uploaderID *uint) (*model.SysUploadFile, error) {
	uploader, err := LoadUploader(application)
	if err != nil {
		return nil, err
	}
	result, err := uploader.Upload(context.Background(), reader, fileName, contentType)
	if err != nil {
		return nil, err
	}
	return saveUploadRecord(application, client, uploaderID, result)
}

// SaveRemoteFile 远程文件抓取：下载远程地址的内容（如微信小程序二维码URL、第三方图片URL），
// 转存到当前默认存储平台并写入记录。文件名取自 URL 路径末段，无扩展名时按 Content-Type 推断。
func SaveRemoteFile(application *app.App, client, remoteURL string, uploaderID *uint) (*model.SysUploadFile, error) {
	parsed, err := validateRemoteURL(remoteURL)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("创建远程文件请求失败: %w", err)
	}
	resp, err := safeRemoteHTTPClient().Do(request)
	if err != nil {
		return nil, fmt.Errorf("获取远程文件失败: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取远程文件失败，状态码: %d", resp.StatusCode)
	}

	// 文件名：URL 路径末段；无扩展名时按 Content-Type 推断
	fileName := path.Base(parsed.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		fileName = fmt.Sprintf("remote_%d", time.Now().UnixMilli())
	}
	if path.Ext(fileName) == "" {
		if ct := resp.Header.Get("Content-Type"); ct != "" {
			if exts, _ := mime.ExtensionsByType(strings.TrimSpace(strings.Split(ct, ";")[0])); len(exts) > 0 {
				fileName += exts[0]
			}
		}
	}

	if resp.ContentLength > MaxUploadSize {
		return nil, errors.New("远程文件大小超过限制（最大50MB）")
	}

	// 响应体直接流入存储层；存储层仍会独立限制最大体积，不能依赖 Content-Length。
	return SaveFileStream(application, client, fileName, resp.Header.Get("Content-Type"), resp.Body, uploaderID)
}
