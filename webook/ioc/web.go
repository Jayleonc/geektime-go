package ioc

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/internal/web/middleware"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx/middleware/ratelimit"
	"github.com/jayleonc/geektime-go/webook/pkg/limiter"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	engine := gin.Default()
	engine.Use(mdls...)

	userHdl.RegisterRoutes(engine)
	wechatHdl.RegisterRoutes(engine)
	return engine
}

func InitGinMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		}),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
		middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
			l.Debug("", logger.Field{Key: "req", Val: al})
		}).AllowRespBody().AllowReqBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).CheckLogin(),
	}
}
