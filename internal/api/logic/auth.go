// Package logic 用户端业务逻辑层。controller 调用 logic，统一传入 *gin.Context。
package logic

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"server_api/internal/api/param"
	"server_api/internal/api/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
	commonupload "server_api/internal/common/upload"
	"server_api/pkg/password"
)

type AuthLogic struct{ App *app.App }

// Login 用户端登录。
func (l *AuthLogic) Login(c *gin.Context, req *param.LoginReq) (*resp.LoginRes, error) {
	var user model.SysUser
	err := l.App.DB.Where("username = ?", req.Username).First(&user).Error
	if err != nil {
		// 区分真实原因，避免把表不存在、数据未导入等问题误报成密码错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, fmt.Errorf("登录异常，请检查数据库: %w", err)
	}
	if !password.Check(user.Password, req.Password) {
		return nil, errors.New("密码错误")
	}
	if user.Status != enums.StatusEnabled {
		return nil, errors.New("账号已被禁用")
	}
	token, err := auth.CreateLogin(l.App, &user, enums.ClientApi, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	return &resp.LoginRes{Token: token}, nil
}

// Logout 退出登录：删除 Redis 中的 login_id，实现 JWT 主动过期。
func (l *AuthLogic) Logout(c *gin.Context) error {
	claims := auth.CtxClaims(c)
	return auth.Logout(l.App, claims.LoginID, enums.LoginStatusLogout)
}

// Profile 当前用户资料。
func (l *AuthLogic) Profile(c *gin.Context) (*resp.ProfileRes, error) {
	claims := auth.CtxClaims(c)
	user, err := auth.GetUserByID(l.App.DB, claims.UserID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	userItem := resp.NewUserItem(*user)
	if userItem.Avatar != "" {
		avatarURL, err := commonupload.FileURL(l.App, userItem.Avatar)
		if err != nil {
			return nil, errors.New("头像地址解析失败")
		}
		userItem.Avatar = avatarURL
	}
	res := &resp.ProfileRes{User: userItem}
	if user.DeptID > 0 {
		var dept model.SysDept
		if err := l.App.DB.First(&dept, user.DeptID).Error; err == nil {
			res.Dept = resp.NewDeptItem(dept)
		}
	}
	return res, nil
}

// ChangePassword 修改自己的密码，成功后踢掉该账号所有在线会话。
func (l *AuthLogic) ChangePassword(c *gin.Context, req *param.ChangePasswordReq) error {
	claims := auth.CtxClaims(c)
	user, err := auth.GetUserByID(l.App.DB, claims.UserID)
	if err != nil {
		return errors.New("用户不存在")
	}
	if !password.Check(user.Password, req.OldPassword) {
		return errors.New("原密码错误")
	}
	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := l.App.DB.Model(user).Update("password", hash).Error; err != nil {
		return errors.New("修改失败")
	}
	return auth.KickUser(l.App, user.ID)
}
