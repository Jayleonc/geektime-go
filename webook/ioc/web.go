package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	"github.com/jayleonc/geektime-go/webook/internal/web/middleware"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx/middleware/ratelimit"
	"github.com/jayleonc/geektime-go/webook/pkg/limiter"
	"github.com/redis/go-redis/v9"
	"time"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl web.UserHandler) *gin.Engine {
	engine := gin.Default()
	engine.Use(mdls...)

	userHdl.RegisterRoutes(engine)
	return engine
}

func InitGinMiddlewares(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			ExposeHeaders: []string{"x-jwt-token"},
		}),
		ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
		(&middleware.LoginJWTMiddlewareBuilder{}).CheckLogin(),
	}
}
