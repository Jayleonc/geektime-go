package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

//	type UserSMSJWTMiddlewareBuilder struct {
//		ijwt.Handler
//	}
//
//	func NewUserSMSJWTMiddlewareBuilder(hdl ijwt.Handler) *UserSMSJWTMiddlewareBuilder {
//		return &UserSMSJWTMiddlewareBuilder{
//			Handler: hdl,
//		}
//	}

var JWTKey = []byte("9dKfy1k348sDkf329skdFjie9120d8fd")

func GenerateToken(tplId string) (string, error) {
	claims := jwt.MapClaims{
		"tpl": tplId,
		"exp": time.Now().Add(time.Hour * 24).Unix(), // 设置过期时间，例如24小时后过期
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JWTKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
