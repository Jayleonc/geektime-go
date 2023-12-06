package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"github.com/jayleonc/geektime-go/webook/internal/service/oauth/wechat"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	uuid "github.com/lithammer/shortuuid/v4"
	"net/http"
)

type OAuth2WechatHandler struct {
	ijwt.Handler
	svc             wechat.Service
	userSvc         service.UserService
	key             []byte
	stateCookieName string
}

func NewOAuth2WechatHandler(svc wechat.Service, userSvc service.UserService, jwtHdl ijwt.Handler) *OAuth2WechatHandler {
	return &OAuth2WechatHandler{
		svc:             svc,
		userSvc:         userSvc,
		key:             []byte("9dKfy1k488Dkf329skdFjiews9098fd"),
		stateCookieName: "jwt-state",
		Handler:         jwtHdl,
	}
}

func (h *OAuth2WechatHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/oauth2/wechat")
	g.GET("/authurl", h.OAuth2URL)
	g.Any("/callback", h.Callback)
}

func (h *OAuth2WechatHandler) OAuth2URL(ctx *gin.Context) {
	state := uuid.New()
	code, err := h.svc.AuthURL(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "构造跳转URL失败",
			Code: 5,
		})
		return
	}
	err = h.setStateToken(ctx, state)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "服务器异常",
			Code: 5,
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: code,
	})
}

func (h *OAuth2WechatHandler) Callback(ctx *gin.Context) {
	err := h.verifyState(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "非法请求",
			Code: 4,
		})
		return
	}
	code := ctx.Query("code")
	wechatInfo, err := h.svc.VerifyCode(ctx, code)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "授权码有误",
			Code: 4,
		})
		return
	}

	u, err := h.userSvc.FindOrCreateByWechat(ctx, wechatInfo)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		return
	}
	err = h.SetLoginToken(ctx, u)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "OK",
	})
	return
}

type StateClaims struct {
	jwt.RegisteredClaims
	State string
}

func (h *OAuth2WechatHandler) setStateToken(ctx *gin.Context, state string) error {
	claims := StateClaims{State: state}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // 1. 生成一个包含 State 的 JWT。
	tokenStr, err := token.SignedString(h.key)                 // 2. 使用预共享的密钥对 JWT 进行签名，以确保其完整性和验证来源。
	if err != nil {
		return err
	}
	// 3.通过设置 Cookie 的方式将 JWT 发送到客户端。
	ctx.SetCookie(h.stateCookieName, tokenStr,
		600, "/oauth2/wechat/callback",
		"", false, true)
	return nil

}

func (h *OAuth2WechatHandler) verifyState(ctx *gin.Context) error {
	state := ctx.Query("state")              // 微信发送回来的 state，如何被 csrf 攻击，这个 state 和 cookie 中的 state 不相等
	ck, err := ctx.Cookie(h.stateCookieName) // 请求参数中的 token，这个 token 存在 cookie 中
	if err != nil {
		return fmt.Errorf("无法获得 cookie %w", err)
	}
	var sc StateClaims
	_, err = jwt.ParseWithClaims(ck, &sc, func(token *jwt.Token) (interface{}, error) {
		return h.key, nil
	})
	if err != nil {
		return fmt.Errorf("解析 token 失败 %w", err)
	}
	if state != sc.State {
		// state 不匹配，有人搞你
		return fmt.Errorf("state 不匹配")
	}
	return nil
}
