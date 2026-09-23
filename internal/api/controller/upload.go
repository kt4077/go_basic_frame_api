package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	commonupload "server_api/internal/common/upload"
	"server_api/pkg/response"
)

type UploadController struct{ App *app.App }

// Upload 通用文件上传（登录即可用）。
func (h *UploadController) Upload(c *gin.Context) {
	claims := auth.CtxClaims(c)
	res, err := commonupload.UploadFile(h.App, c, enums.ClientApi, &claims.UserID)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	result, err := commonupload.NewFileResponse(h.App, res)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, result)
}
