package jwt

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

var (
	JWTKey   = []byte("9dKfy1k348sDkf329skdFjiews90d8fd")
	RCJWTKey = []byte("9dKfy1k348sDkf329skdFjiews90d8ff")
)

type RedisJWTHandler struct {
	client        redis.Cmdable
	signingMethod jwt.SigningMethod
	rcExpiration  time.Duration
}

func NewRedisJWTHandler(client redis.Cmdable) Handler {
	return &RedisJWTHandler{
		client:        client,
		signingMethod: jwt.SigningMethodHS512,
		rcExpiration:  time.Hour * 24 * 7,
	}
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Ssid      string
	Email     string
	Phone     string
	UserAgent string
}

func (h *RedisJWTHandler) ClearToken(ctx *gin.Context) error {
	ctx.Header("x-jwt-token", "")
	ctx.Header("x-refresh-token", "")
	user := ctx.MustGet("user").(UserClaims)
	return h.client.Set(ctx, fmt.Sprintf("user:ssid:%s", user.Ssid), "", h.rcExpiration).Err()
}

// ExtractToken 根据约定，token 在 Authorization 头部
// Bearer XXXX
func (h *RedisJWTHandler) ExtractToken(ctx *gin.Context) string {
	authCode := ctx.GetHeader("Authorization")
	if authCode == "" {
		return authCode
	}
	segs := strings.Split(authCode, " ")
	if len(segs) != 2 {
		return ""
	}
	return segs[1]
}

func (h *RedisJWTHandler) SetLoginToken(ctx *gin.Context, u domain.User) error {
	ssid := uuid.New().String()
	err := h.setRefreshToken(ctx, u, ssid)
	if err != nil {
		return err
	}
	return h.SetJWTToken(ctx, u, ssid)
}

// SetJWTToken 短 token 在正常业务请求中使用
func (h *RedisJWTHandler) SetJWTToken(ctx *gin.Context, u domain.User, ssid string) error {
	uc := UserClaims{
		Uid:       u.Id,
		Ssid:      ssid,
		Email:     u.Email,
		Phone:     u.Phone,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-jwt-token", tokenStr)
	return nil
}

// setRefreshToken 长 token 只在登录和调用 web.UserHandler{}.RefreshToken 时使用，不容易泄漏
func (h *RedisJWTHandler) setRefreshToken(ctx *gin.Context, u domain.User, ssid string) error {
	uc := UserClaims{
		Uid:       u.Id,
		Ssid:      ssid,
		Email:     u.Email,
		Phone:     u.Phone,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.rcExpiration)),
		},
	}
	token := jwt.NewWithClaims(h.signingMethod, uc)
	tokenStr, err := token.SignedString(RCJWTKey)
	if err != nil {
		return err
	}
	ctx.Header("x-refresh-token", tokenStr)
	return nil
}

func (h *RedisJWTHandler) CheckSession(ctx *gin.Context, ssid string) error {
	fmt.Println("CheckSession")
	cnt, err := h.client.Exists(ctx, fmt.Sprintf("user:ssid:%s", ssid)).Result()
	// 如果 redis 服务有问题时，想正常提供服务，可以注释对 err 的判断
	if err != nil {
		return err
	}
	// token 无效
	if cnt > 0 {
		return errors.New("token 无效")
	}
	return nil
}
