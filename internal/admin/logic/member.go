package logic

import (
	"errors"

	"github.com/gin-gonic/gin"

	"server_api/internal/admin/param"
	"server_api/internal/admin/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	commonupload "server_api/internal/common/upload"
	"server_api/pkg/pagination"
)

type MemberLogic struct{ App *app.App }

// List 会员用户分页列表，支持关键字、账号状态、注册来源过滤。
func (l *MemberLogic) List(c *gin.Context, req *param.MemberListReq) (*resp.MemberListRes, error) {
	var total int64
	var members []model.SysMember
	db := l.App.DB.Model(&model.SysMember{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		db = db.Where("sn LIKE ? OR nickname LIKE ? OR real_name LIKE ? OR account LIKE ? OR mobile LIKE ?",
			like, like, like, like, like)
	}
	if req.Status != 0 {
		db = db.Where("status = ?", req.Status)
	}
	if req.RegisterSource != 0 {
		db = db.Where("register_source = ?", req.RegisterSource)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	page, pageSize := pagination.Normalize(req.Page, req.PageSize)
	if err := db.Order("id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&members).Error; err != nil {
		return nil, errors.New("查询失败")
	}
	items := resp.NewMemberItems(members)
	if err := completeMemberAvatarList(l.App, items); err != nil {
		return nil, errors.New("头像地址解析失败")
	}
	return &resp.MemberListRes{List: items, Total: total}, nil
}

// SetStatus 启用/禁用会员账号。
func (l *MemberLogic) SetStatus(c *gin.Context, req *param.MemberSetStatusReq) error {
	var member model.SysMember
	if err := l.App.DB.First(&member, req.ID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if member.Status == req.Status {
		return nil
	}
	if err := l.App.DB.Model(&member).Update("status", req.Status).Error; err != nil {
		return errors.New("操作失败")
	}
	// 禁用后立即使其在线会话失效
	if req.Status == enums.StatusDisabled {
		return auth.KickMember(l.App, &member)
	}
	return nil
}

func completeMemberAvatarList(application *app.App, list []resp.MemberItem) error {
	resolver, err := commonupload.NewURLResolver(application)
	if err != nil {
		for _, item := range list {
			if item.Avatar != "" {
				return err
			}
		}
		return nil
	}
	for i := range list {
		list[i].Avatar = resolver.URL(list[i].Avatar)
	}
	return nil
}
