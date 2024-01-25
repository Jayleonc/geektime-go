package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	ijwt "github.com/jayleonc/geektime-go/webook/internal/web/jwt"
	"github.com/jayleonc/geektime-go/webook/internal/web/vo"
	"github.com/jayleonc/geektime-go/webook/pkg/ginx"
	"net/http"
	"time"
)

var (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRegexp    *regexp.Regexp
	passwordRegexp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService, jwtHdl ijwt.Handler) *UserHandler {
	return &UserHandler{
		emailRegexp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		Handler:        jwtHdl,
		svc:            svc,
		codeSvc:        codeSvc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WrapBody(h.SignUp))
	ug.POST("/login", ginx.WrapBody(h.LoginJWT))
	ug.POST("/logout", h.LogoutJWT)
	ug.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))
	ug.GET("/profile", ginx.WarpClaims(h.Profile))

	// 手机验证码登录相关功能
	ug.POST("/login_sms/code/send", ginx.WrapBody(h.SendSMSLoginCode))
	ug.POST("/login_sms", ginx.WrapBody(h.LoginSMS))
	ug.POST("/refresh_token", h.RefreshToken)

}

func (h *UserHandler) SignUp(ctx *gin.Context, req vo.UserSignUpReq) (ginx.Response, error) {
	isEmail, err := h.emailRegexp.MatchString(req.Email)
	if err != nil {
		return ginx.Response{Code: 200, Msg: "系统错误"}, err
	}

	if !isEmail {
		return ginx.Response{Code: 200, Msg: "邮箱格式错误"}, nil
	}

	isPassword, err := h.passwordRegexp.MatchString(req.Password)
	if err != nil {
		return ginx.Response{Code: 200, Msg: "系统错误"}, err
	}

	if !isPassword {
		return ginx.Response{Code: 200, Msg: "密码格式错误"}, nil
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    req.Email,
		Password: req.Password,
	})

	switch err {
	case nil:
		return ginx.Response{Msg: "注册成功"}, nil
	case service.ErrDuplicateEmail:
		return ginx.Response{Code: 200, Msg: "邮箱冲突，请换一个"}, err
	default:
		return ginx.Response{Code: 200, Msg: "系统错误"}, err
	}

}
func (h *UserHandler) LoginJWT(ctx *gin.Context, req vo.UserLoginReq) (ginx.Response, error) {

	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		err = h.SetLoginToken(ctx, u)
		if err != nil {
			return ginx.Response{Code: 200, Msg: "系统错误"}, err
		}
		return ginx.Response{Msg: "登陆成功"}, nil
	case service.ErrInvalidUserOrPassword:
		return ginx.Response{Code: http.StatusOK, Msg: "用户名或密码不对"}, err
	default:
		return ginx.Response{Code: http.StatusOK, Msg: "系统错误"}, err
	}
}

func (h *UserHandler) Edit(ctx *gin.Context, req vo.UserEditReq, uc ijwt.UserClaims) (ginx.Response, error) {
	var u = domain.User{
		Id:       uc.Uid,
		Nickname: req.Nickname,
		AboutMe:  req.AboutMe,
	}

	err := h.svc.Update(ctx, u)
	if err != nil {
		return ginx.Response{
			Code: http.StatusInternalServerError,
			Msg:  "系统错误",
		}, err
	}

	return ginx.Response{
		Msg: "OK",
	}, nil
}

func (h *UserHandler) Profile(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Response, error) {
	u, err := h.svc.FindById(ctx, uc.Uid)
	if err != nil {
		return ginx.Response{Code: 200, Msg: "系统异常"}, err
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
	}
	return ginx.Response{Data: User{
		Nickname: u.Nickname,
		Email:    u.Email,
		Phone:    u.Phone,
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.Format(time.DateOnly),
	}}, nil
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context, req vo.SendSMSLoginReq) (ginx.Response, error) {
	if req.Phone == "" {
		return ginx.Response{Code: 4, Msg: "请输入手机号"}, errors.New("用户输入手机号为空")
	}
	err := h.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch {
	case err == nil:
		return ginx.Response{
			Msg: "发送成功",
		}, nil
	case errors.Is(err, service.ErrCodeSendTooMany):
		return ginx.Response{Code: 400, Msg: "短信发送太频繁，请稍后再试"}, err
	default:
		return ginx.Response{Code: 500, Msg: "系统错误"}, err
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req vo.LoginSMSReq) (ginx.Response, error) {
	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		return ginx.Response{
			Code: 500,
			Msg:  "系统错误",
		}, errors.New("手机验证码验证失败, " + err.Error())
	}
	if !ok {
		return ginx.Response{
			Code: 400,
			Msg:  "验证码错误，请重新输入",
		}, errors.New("验证码错误")
	}

	user, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return ginx.Response{
			Code: 500,
			Msg:  "系统错误",
		}, err
	}
	err = h.SetLoginToken(ctx, user)
	if err != nil {
		return ginx.Response{
			Msg: "系统错误",
		}, err
	}
	return ginx.Response{Msg: "登录成功"}, nil
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	extractToken := h.ExtractToken(ctx) // 解析出临时短 token
	var rc ijwt.UserClaims
	token, err := jwt.ParseWithClaims(extractToken, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	err = h.SetJWTToken(ctx, domain.User{
		Id:    rc.Uid,
		Email: rc.Email,
		Phone: rc.Phone,
	}, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, ginx.Response{
		Msg: "OK",
	})
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Response{
			Code: 5,
			Msg:  "系统错误",
		})
		return
	}
	ctx.JSON(http.StatusOK, ginx.Response{
		Msg: "退出登录成功 ",
	})
}

func (h *UserHandler) Login(ctx *gin.Context) {
	type req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var r req
	if err := ctx.Bind(&r); err != nil {
		return
	}

	u, err := h.svc.Login(ctx, r.Email, r.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{MaxAge: 900})
		if err = sess.Save(); err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登陆成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}
