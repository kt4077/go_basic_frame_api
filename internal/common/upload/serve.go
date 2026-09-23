package upload

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
	"server_api/internal/common/enums"
)

// ServeLocalFile 根据当前默认本地存储配置提供文件访问。
// 路径必须是上传模块生成的 uploads/... 对象Key，并经过目录穿越校验。
func ServeLocalFile(application *app.App, c *gin.Context) {
	cfg, err := LoadDefaultStorage(application)
	if err != nil || cfg.Channel != enums.StorageChannelLocal {
		c.Status(http.StatusNotFound)
		return
	}
	params, err := StorageParams(cfg)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	key := strings.TrimLeft(strings.ReplaceAll(c.Param("filepath"), "\\", "/"), "/")
	if !strings.HasPrefix(key, "uploads/") {
		c.Status(http.StatusNotFound)
		return
	}
	root := params["root_path"]
	if root == "" {
		root = "./uploads"
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	target := filepath.Join(rootAbs, filepath.FromSlash(strings.TrimPrefix(key, "uploads/")))
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		c.Status(http.StatusForbidden)
		return
	}
	c.File(targetAbs)
}
