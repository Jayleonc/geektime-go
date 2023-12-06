package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"net/http"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			return
		}
		// 检查头部 Authorization
		tokenStr := m.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})

		if err != nil || !token.Valid {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}

		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set("user", uc)
		//if uc.UserAgent != ctx.GetHeader("User-Agent") {
		//	// todo 监控告警时，需要记录这里的发生
		//	// 进来这里的，大概率是攻击者
		//	fmt.Println("User-Agent 出错啦！")
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}

		//expireTime := uc.ExpiresAt
		//if expireTime.Sub(time.Now()) < time.Minute*15 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
		//	tokenStr, err = token.SignedString(web.JWTKey)
		//	ctx.Header("x-jwt-token", tokenStr)
		//	fmt.Println("-------- token 刷新啦 ------")
		//	if err != nil {
		//		fmt.Println(err)
		//	}
		//}
	}
}
