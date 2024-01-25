package ginx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jayleonc/geektime-go/webook/pkg/logger"
	"net/http"
)

var L logger.Logger = logger.NewNopLogger()

func WrapBodyAndClaims[Req any, Claims jwt.Claims](bizFn func(ctx *gin.Context, req Req, uc Claims) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			L.Error("输入错误", logger.Error(err))
			return
		}
		L.Debug("输入参数", logger.Field{Key: "req", Val: req})
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, req, uc)
		if err != nil {
			L.Error("执行业务逻辑失败", logger.Error(err))
		}
		Return(ctx, res, err)
	}
}

func WrapBody[Req any](bizFn func(ctx *gin.Context, req Req) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.Bind(&req); err != nil {
			L.Error("输入错误", logger.Error(err))
			return
		}
		L.Debug("输入参数", logger.Field{Key: "req", Val: req})
		res, err := bizFn(ctx, req)
		if err != nil {
			L.Error("执行业务逻辑失败", logger.Error(err))
		}
		Return(ctx, res, err)
	}
}

func WarpClaims[Claims any](bizFn func(ctx *gin.Context, uc Claims) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		val, ok := ctx.Get("user")
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		uc, ok := val.(Claims)
		if !ok {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		res, err := bizFn(ctx, uc)
		if err != nil {
			L.Error("执行业务逻辑失败", logger.Error(err))
		}
		Return(ctx, res, err)
	}
}

func Wrap(fn func(ctx *gin.Context) (Response, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res, err := fn(ctx)
		if err != nil {
			L.Error("处理业务逻辑出错",
				logger.String("path", ctx.Request.URL.Path),
				logger.String("route", ctx.FullPath()),
				logger.Error(err),
			)
		}
		Return(ctx, res, err)
	}
}
