package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Time{})
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()

		const updateTimeKey = "update_time"
		// 获取上一次刷新的时间
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time)
		if !ok || now.Sub(lastUpdateTime) > time.Second*10 {
			// 第一次进来或需要刷新
			sess.Set(updateTimeKey, now)
			sess.Set("userId", userId)
			if err := sess.Save(); err != nil {
				fmt.Println(err)
			}
		}

	}
}
