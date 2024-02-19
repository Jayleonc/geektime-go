package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jayleonc/geektime-go/webook/internal/web"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/internal/web/middleware"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx/middleware/prometheus"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func InitWebServer(mdls []gin.HandlerFunc, userHdl *web.UserHandler, wechatHdl *web.OAuth2WechatHandler, artHdl *web.ArticleHandler) *gin.Engine {
	engine := gin.Default()
	engine.Use(mdls...)

	userHdl.RegisterRoutes(engine)
	wechatHdl.RegisterRoutes(engine)
	artHdl.RegisterRoutes(engine)
	return engine
}

func InitGinMiddlewares(redisClient redis.Cmdable, jwtHdl ijwt.Handler, l logger.Logger) []gin.HandlerFunc {
	pb := prometheus.Builder{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "gin_http",
		Help:      "统计 Gin 的 http 请求",
	}
	ginx.InitCounter(prometheus2.CounterOpts{
		Namespace: "geektime_jayleonc",
		Subsystem: "webook",
		Name:      "biz_code",
		Help:      "统计业务错误码",
	})
	return []gin.HandlerFunc{
		cors.New(cors.Config{
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return true
			},
			ExposeHeaders: []string{"x-jwt-token", "x-refresh-token"},
		}),
		//ratelimit.NewBuilder(limiter.NewRedisSlidingWindowLimiter(redisClient, time.Second, 1000)).Build(),
		//middleware.NewLogMiddlewareBuilder(func(ctx context.Context, al middleware.AccessLog) {
		//	l.Debug("", logger.Field{Key: "req", Val: al})
		//}).AllowRespBody().AllowReqBody().Build(),
		middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).CheckLogin(),
		pb.BuilderResponseTime(),
		pb.BuilderActiveRequest(),
		otelgin.Middleware("webook"),
	}
}
