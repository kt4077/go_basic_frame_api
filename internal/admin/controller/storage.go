package controller

import (
	"github.com/gin-gonic/gin"

	"server_api/internal/admin/logic"
	"server_api/internal/admin/param"
	"server_api/pkg/response"
	requestvalidate "server_api/pkg/validate"
)

type StorageController struct{ Logic *logic.StorageLogic }

// List 存储渠道列表。
func (h *StorageController) List(c *gin.Context) {
	res, err := h.Logic.List(c)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Create 新增存储渠道。
func (h *StorageController) Create(c *gin.Context) {
	var req param.StorageSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	res, err := h.Logic.Create(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// Update 修改存储渠道。
func (h *StorageController) Update(c *gin.Context) {
	var req param.StorageSaveReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if req.ID == 0 {
		response.Fail(c, response.CodeErrParams, "主键ID不能为空")
		return
	}
	res, err := h.Logic.Update(c, &req)
	if err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, res)
}

// SetDefault 设为默认渠道。
func (h *StorageController) SetDefault(c *gin.Context) {
	var req param.IDReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if err := h.Logic.SetDefault(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}

// Delete 删除存储渠道。
func (h *StorageController) Delete(c *gin.Context) {
	var req param.IDReq
	if err := requestvalidate.Bind(c, &req); err != nil {
		response.Fail(c, response.CodeErrParams, err.Error())
		return
	}
	if err := h.Logic.Delete(c, &req); err != nil {
		response.Fail(c, response.CodeErrBusiness, err.Error())
		return
	}
	response.OK(c, nil)
}
