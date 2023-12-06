package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
)

type Handler interface {
	ClearToken(ctx *gin.Context) error
	ExtractToken(ctx *gin.Context) string
	SetLoginToken(ctx *gin.Context, user domain.User) error
	SetJWTToken(ctx *gin.Context, user domain.User, ssid string) error
	CheckSession(ctx *gin.Context, ssid string) error
}
