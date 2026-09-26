package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"server_api/internal/common/app"
	"server_api/internal/common/enums"
	"server_api/internal/common/model"
)

// JWT Claims：除常规字段外携带 login_id，用于 Redis 主动过期。
// 会员令牌额外携带昵称、用户编号与头像相对路径，鉴权通过后写入上下文。
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	LoginID  string `json:"login_id"`
	Nickname string `json:"nickname"`
	SN       string `json:"sn"`
	Avatar   string `json:"avatar"`
	IsSuper  int    `json:"is_super"`
	Client   string `json:"client"` // admin | api
	jwt.RegisteredClaims
}

// loginCacheKey Redis 中 login_id 的缓存键。
func loginCacheKey(loginID string) string { return "login:" + loginID }

// memberLoginCacheKey Redis 中会员当前会话的缓存键，按会员ID存放 login_id。
func memberLoginCacheKey(memberID uint) string {
	return "member_login:" + strconv.FormatUint(uint64(memberID), 10)
}

// CreateLogin 登录成功后调用：写入登录流水、生成 JWT、写入 Redis 缓存。
func CreateLogin(application *app.App, user *model.SysUser, client, ip, userAgent string) (token string, err error) {
	loginID := uuid.NewString()
	record := &model.SysUserLogin{
		LoginID:   loginID,
		UserID:    user.ID,
		Username:  user.Username,
		Client:    client,
		LoginIP:   ip,
		UserAgent: userAgent,
		LoginAt:   time.Now(),
		Status:    enums.LoginStatusOnline,
	}
	if err = application.DB.Create(record).Error; err != nil {
		return "", err
	}

	expire := time.Duration(application.Cfg.Jwt.ExpireHours) * time.Hour
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		LoginID:  loginID,
		IsSuper:  user.IsSuper,
		Client:   client,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    application.Cfg.Jwt.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(application.Cfg.Jwt.Secret))
	if err != nil {
		return "", err
	}

	// login_id 缓存：存在即有效；退出/禁用/踢人时删除，实现 JWT 主动过期
	if err = application.Redis.Set(context.Background(), loginCacheKey(loginID), user.ID, expire).Err(); err != nil {
		return "", err
	}
	return token, nil
}

// Logout 主动退出：更新流水状态并删除 Redis 中的 login_id 缓存。
func Logout(application *app.App, loginID string, status int) error {
	now := time.Now()
	err := application.DB.Model(&model.SysUserLogin{}).
		Where("login_id = ? AND status = ?", loginID, enums.LoginStatusOnline).
		Updates(map[string]interface{}{"status": status, "logout_at": now}).Error
	if err != nil {
		return err
	}
	return application.Redis.Del(context.Background(), loginCacheKey(loginID)).Err()
}

// KickUser 踢人/禁用：删除该用户当前在线的所有 login_id 缓存。
func KickUser(application *app.App, userID uint) error {
	var records []model.SysUserLogin
	if err := application.DB.Where("user_id = ? AND status = ?", userID, enums.LoginStatusOnline).
		Find(&records).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := application.DB.Model(&model.SysUserLogin{}).
		Where("user_id = ? AND status = ?", userID, enums.LoginStatusOnline).
		Updates(map[string]interface{}{"status": enums.LoginStatusKicked, "logout_at": now}).Error; err != nil {
		return err
	}
	ctx := context.Background()
	for _, r := range records {
		application.Redis.Del(ctx, loginCacheKey(r.LoginID))
	}
	return nil
}

// ParseToken 解析并校验 JWT，同时检查 Redis 中 login_id 是否仍存在（主动过期）。
func ParseToken(application *app.App, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非法的签名方式")
		}
		return []byte(application.Cfg.Jwt.Secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("token 无效")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("token 无效")
	}
	// JWT 自身未过期，但 Redis 缓存已被删除（退出/禁用/踢人），同样视为失效
	exists, err := application.Redis.Exists(context.Background(), loginCacheKey(claims.LoginID)).Result()
	if err != nil {
		return nil, errors.New("缓存服务异常")
	}
	if exists == 0 {
		return nil, errors.New("登录已失效，请重新登录")
	}
	// 会员令牌仅允许当前会话：重新登录后 member_login 缓存被覆盖，旧令牌失效
	if claims.Client == enums.ClientApi {
		current, err := application.Redis.Get(context.Background(), memberLoginCacheKey(claims.UserID)).Result()
		if err != nil {
			return nil, errors.New("登录已失效，请重新登录")
		}
		if current != claims.LoginID {
			return nil, errors.New("登录已失效，请重新登录")
		}
	}
	return claims, nil
}

// 从请求头提取 Bearer token。
func BearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// 当前登录用户相关的 context key。
const (
	CtxClaimsKey = "claims"
)

// CtxClaims 从 context 中取登录信息。
func CtxClaims(c *gin.Context) *Claims {
	v, ok := c.Get(CtxClaimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*Claims)
	return claims
}

// GetUserByID 按 ID 查询用户。关联数据由各模块的响应结构负责组装。
func GetUserByID(db *gorm.DB, id uint) (*model.SysUser, error) {
	var user model.SysUser
	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateMemberLogin 会员登录成功后调用：更新会员会话与最近登录信息，
// 生成携带昵称/编号/头像的 JWT，并写入 Redis 会话缓存。
// 会员单端登录：member_login 缓存按会员ID覆盖，旧令牌随之失效。
func CreateMemberLogin(application *app.App, member *model.SysMember, ip, userAgent string) (string, error) {
	loginID := uuid.NewString()
	expire := time.Duration(application.Cfg.Jwt.ExpireHours) * time.Hour
	claims := &Claims{
		UserID:   member.ID,
		Username: memberUsername(member),
		LoginID:  loginID,
		Nickname: member.Nickname,
		SN:       member.SN,
		Avatar:   member.Avatar,
		Client:   enums.ClientApi,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    application.Cfg.Jwt.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(application.Cfg.Jwt.Secret))
	if err != nil {
		return "", err
	}

	now := time.Now()
	if err := application.DB.Model(&model.SysMember{}).Where("id = ?", member.ID).
		Updates(map[string]interface{}{"login_id": loginID, "login_ip": ip, "logged_at": now}).Error; err != nil {
		return "", err
	}

	ctx := context.Background()
	pipe := application.Redis.TxPipeline()
	pipe.Set(ctx, loginCacheKey(loginID), member.ID, expire)
	pipe.Set(ctx, memberLoginCacheKey(member.ID), loginID, expire)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", err
	}
	return token, nil
}

// LogoutMember 会员退出登录：删除会话缓存。
func LogoutMember(application *app.App, memberID uint, loginID string) error {
	ctx := context.Background()
	return application.Redis.Del(ctx, loginCacheKey(loginID), memberLoginCacheKey(memberID)).Err()
}

// KickMember 禁用会员时调用：删除其当前会话缓存，令牌立即失效。
func KickMember(application *app.App, member *model.SysMember) error {
	if member.LoginID == "" {
		return nil
	}
	ctx := context.Background()
	return application.Redis.Del(ctx, loginCacheKey(member.LoginID), memberLoginCacheKey(member.ID)).Err()
}

// memberUsername 会员令牌展示名：优先登录账号，未设置时使用手机号。
func memberUsername(member *model.SysMember) string {
	if member.Account != nil && *member.Account != "" {
		return *member.Account
	}
	return member.Mobile
}

// ClientAllowed 判断令牌客户端类型是否在允许列表内；未指定时放行全部类型。
func ClientAllowed(claims *Claims, allow []string) bool {
	if len(allow) == 0 {
		return true
	}
	for _, client := range allow {
		if claims.Client == client {
			return true
		}
	}
	return false
}
