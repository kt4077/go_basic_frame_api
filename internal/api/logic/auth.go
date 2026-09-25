// Package logic 用户端业务逻辑层。controller 调用 logic，统一传入 *gin.Context。
// 会员认证相关业务：短信验证码、注册、登录、资料与账号安全设置。
package logic

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"server_api/internal/api/param"
	"server_api/internal/api/resp"
	"server_api/internal/common/app"
	"server_api/internal/common/auth"
	commonmiddleware "server_api/internal/common/middleware"
	"server_api/internal/common/model"
	commonupload "server_api/internal/common/upload"
	"server_api/pkg/dberror"
	"server_api/pkg/mask"
	"server_api/pkg/password"
	"server_api/pkg/sn"
)

type AuthLogic struct {
	App *app.App
	SMS *SmsLogic
}

// 账号仅允许字母、数字、下划线，长度 4-32；手机号固定 11 位、1 开头。
var (
	mobilePattern  = regexp.MustCompile(`^1\d{10}$`)
	accountPattern = regexp.MustCompile(`^\w{4,32}$`)
)

// Register 手机号验证码注册：默认资料取用户端平台配置，注册成功后直接登录。
func (l *AuthLogic) Register(c *gin.Context, req *param.RegisterReq) (*resp.LoginRes, error) {
	if !mobilePattern.MatchString(req.Mobile) {
		return nil, errors.New("手机号格式不正确")
	}
	if err := l.SMS.VerifyCode(smsSceneRegister, req.Mobile, req.Code); err != nil {
		return nil, err
	}
	var count int64
	l.App.DB.Model(&model.SysMember{}).Where("mobile = ?", req.Mobile).Count(&count)
	if count > 0 {
		return nil, errors.New("该手机号已注册，请直接登录")
	}

	nickname, avatar := l.defaultMemberInfo()
	now := time.Now()
	// 用户编号唯一：生成失败或撞号时重试
	for i := 0; i < 5; i++ {
		memberSN, snErr := sn.Generate()
		if snErr != nil {
			return nil, errors.New("注册失败，请稍后重试")
		}
		member := &model.SysMember{
			SN:             memberSN,
			Nickname:       nickname,
			Avatar:         avatar,
			Mobile:         req.Mobile,
			RegisterIP:     c.ClientIP(),
			RegisteredAt:   &now,
			RegisterSource: commonmiddleware.CtxPlatformSource(c),
			Status:         1,
			Balance:        "0.00",
		}
		err := l.App.DB.Create(member).Error
		if err == nil {
			return l.issueToken(c, member)
		}
		if dberror.IsDuplicateKey(err) && strings.Contains(err.Error(), "sn") {
			continue
		}
		if dberror.IsDuplicateKey(err) {
			return nil, errors.New("该手机号已注册，请直接登录")
		}
		return nil, errors.New("注册失败，请稍后重试")
	}
	return nil, errors.New("注册失败，请稍后重试")
}

// LoginPassword 账号密码登录。账号不存在时明确提示，便于用户区分注册与登录。
func (l *AuthLogic) LoginPassword(c *gin.Context, req *param.PasswordLoginReq) (*resp.LoginRes, error) {
	account := strings.TrimSpace(req.Account)
	var member model.SysMember
	err := l.App.DB.Where("account = ?", account).First(&member).Error
	if err != nil {
		return nil, errors.New("账号不存在")
	}
	if member.Password == "" {
		return nil, errors.New("该账号未设置密码，请使用短信验证码登录")
	}
	if !password.Check(member.Password, req.Password) {
		return nil, errors.New("密码错误")
	}
	return l.loginEnabled(c, &member)
}

// LoginSms 手机号验证码登录。
func (l *AuthLogic) LoginSms(c *gin.Context, req *param.SmsLoginReq) (*resp.LoginRes, error) {
	if !mobilePattern.MatchString(req.Mobile) {
		return nil, errors.New("手机号格式不正确")
	}
	if err := l.SMS.VerifyCode(smsSceneLogin, req.Mobile, req.Code); err != nil {
		return nil, err
	}
	var member model.SysMember
	if err := l.App.DB.Where("mobile = ?", req.Mobile).First(&member).Error; err != nil {
		return nil, errors.New("账号不存在，请先注册")
	}
	return l.loginEnabled(c, &member)
}

// loginEnabled 校验账号状态后签发令牌。
func (l *AuthLogic) loginEnabled(c *gin.Context, member *model.SysMember) (*resp.LoginRes, error) {
	if member.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	return l.issueToken(c, member)
}

// issueToken 生成登录令牌并刷新最近登录信息。
func (l *AuthLogic) issueToken(c *gin.Context, member *model.SysMember) (*resp.LoginRes, error) {
	token, err := auth.CreateMemberLogin(l.App, member, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		return nil, errors.New("登录失败，请稍后重试")
	}
	return &resp.LoginRes{Token: token}, nil
}

// defaultMemberInfo 读取用户端平台配置中的默认昵称与默认头像。
func (l *AuthLogic) defaultMemberInfo() (string, string) {
	var config model.SysPlatformConfig
	err := l.App.DB.Where("type = ?", 2).First(&config).Error
	nickname := config.DefaultNickname
	if err != nil || strings.TrimSpace(nickname) == "" {
		return "", ""
	}
	return nickname, config.DefaultAvatar
}

// Logout 退出登录：删除当前会话缓存。
func (l *AuthLogic) Logout(c *gin.Context) error {
	claims := auth.CtxClaims(c)
	return auth.LogoutMember(l.App, claims.UserID, claims.LoginID)
}

// Profile 当前会员资料。
func (l *AuthLogic) Profile(c *gin.Context) (*resp.MemberProfile, error) {
	claims := auth.CtxClaims(c)
	member, err := l.currentMember(claims)
	if err != nil {
		return nil, err
	}
	item, err := l.toProfile(member)
	if err != nil {
		return nil, errors.New("头像地址解析失败")
	}
	return item, nil
}

// UpdateProfile 修改昵称、姓名与头像。
func (l *AuthLogic) UpdateProfile(c *gin.Context, req *param.ProfileUpdateReq) (*resp.MemberProfile, error) {
	claims := auth.CtxClaims(c)
	avatar, err := commonupload.NormalizeFilePath(l.App, req.Avatar)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}
	if len(updates) == 0 {
		return nil, errors.New("请填写需要修改的内容")
	}
	if err := l.App.DB.Model(&model.SysMember{}).Where("id = ?", claims.UserID).Updates(updates).Error; err != nil {
		return nil, errors.New("修改失败")
	}
	member, err := l.currentMember(claims)
	if err != nil {
		return nil, err
	}
	return l.toProfile(member)
}

// UpdateAccount 修改登录账号，账号全表唯一。
func (l *AuthLogic) UpdateAccount(c *gin.Context, req *param.AccountUpdateReq) error {
	account := strings.TrimSpace(req.Account)
	if !accountPattern.MatchString(account) {
		return errors.New("登录账号需为4-32位字母、数字或下划线")
	}
	claims := auth.CtxClaims(c)
	var count int64
	l.App.DB.Model(&model.SysMember{}).Where("account = ?", account).Count(&count)
	if count > 0 {
		return errors.New("该登录账号已被使用")
	}
	err := l.App.DB.Model(&model.SysMember{}).Where("id = ?", claims.UserID).
		Update("account", account).Error
	if err != nil {
		if dberror.IsDuplicateKey(err) {
			return errors.New("该登录账号已被使用")
		}
		return errors.New("修改失败")
	}
	return nil
}

// UpdatePassword 修改登录密码：强度要求6位以上且同时包含字母和数字，
// 已设置过密码时必须校验原密码；修改成功后当前会话立即失效需重新登录。
func (l *AuthLogic) UpdatePassword(c *gin.Context, req *param.PasswordUpdateReq) error {
	if !validPasswordStrength(req.NewPassword) {
		return errors.New("密码至少6位，且必须同时包含字母和数字")
	}
	claims := auth.CtxClaims(c)
	member, err := l.currentMember(claims)
	if err != nil {
		return err
	}
	if member.Password != "" {
		if req.OldPassword == "" {
			return errors.New("请输入原密码")
		}
		if !password.Check(member.Password, req.OldPassword) {
			return errors.New("原密码错误")
		}
	}
	hash, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}
	if err := l.App.DB.Model(&model.SysMember{}).Where("id = ?", claims.UserID).
		Update("password", hash).Error; err != nil {
		return errors.New("修改失败")
	}
	return auth.KickMember(l.App, member)
}

// UpdateMobile 修改手机号：需接收新手机号短信验证码，且新手机号未被绑定。
func (l *AuthLogic) UpdateMobile(c *gin.Context, req *param.MobileUpdateReq) error {
	if !mobilePattern.MatchString(req.Mobile) {
		return errors.New("手机号格式不正确")
	}
	if err := l.SMS.VerifyCode(smsSceneChangeMobile, req.Mobile, req.Code); err != nil {
		return err
	}
	claims := auth.CtxClaims(c)
	var count int64
	l.App.DB.Model(&model.SysMember{}).Where("mobile = ?", req.Mobile).Count(&count)
	if count > 0 {
		return errors.New("该手机号已绑定其他账号")
	}
	if err := l.App.DB.Model(&model.SysMember{}).Where("id = ?", claims.UserID).
		Update("mobile", req.Mobile).Error; err != nil {
		if dberror.IsDuplicateKey(err) {
			return errors.New("该手机号已绑定其他账号")
		}
		return errors.New("修改失败")
	}
	return nil
}

// currentMember 从数据库加载当前登录会员，禁用后即使令牌未过期也立即失效。
func (l *AuthLogic) currentMember(claims *auth.Claims) (*model.SysMember, error) {
	var member model.SysMember
	if err := l.App.DB.First(&member, claims.UserID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	if member.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}
	return &member, nil
}

// toProfile 组装会员资料，头像按当前存储配置补全，手机号脱敏展示。
func (l *AuthLogic) toProfile(member *model.SysMember) (*resp.MemberProfile, error) {
	avatar, err := commonupload.FileURL(l.App, member.Avatar)
	if err != nil {
		return nil, err
	}
	account := ""
	if member.Account != nil {
		account = *member.Account
	}
	registeredAt := time.Time{}
	if member.RegisteredAt != nil {
		registeredAt = *member.RegisteredAt
	}
	return &resp.MemberProfile{
		ID:             member.ID,
		SN:             member.SN,
		Account:        account,
		Nickname:       member.Nickname,
		RealName:       member.RealName,
		Avatar:         avatar,
		Mobile:         mask.Mobile(member.Mobile),
		Gender:         member.Gender,
		RegisterSource: member.RegisterSource,
		RegisteredAt:   registeredAt,
	}, nil
}

// validPasswordStrength 密码强度：至少6位，且同时包含字母和数字。
func validPasswordStrength(value string) bool {
	if len(value) < 6 {
		return false
	}
	hasLetter, hasDigit := false, false
	for _, ch := range value {
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z':
			hasLetter = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}
