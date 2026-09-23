package auth

import (
	"context"
	"errors"
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
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	LoginID  string `json:"login_id"`
	IsSuper  int    `json:"is_super"`
	Client   string `json:"client"` // admin | api
	jwt.RegisteredClaims
}

// loginCacheKey Redis 中 login_id 的缓存键。
func loginCacheKey(loginID string) string { return "login:" + loginID }

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
