package web

import (
	"errors"
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/jayleonc/geektime-go/webook/internal/domain"
	"github.com/jayleonc/geektime-go/webook/internal/service"
	"net/http"
	"time"
)

var (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	JWTKey               = []byte("9dKfy1k348sDkf329skdFjiews90d8fd")
	bizLogin             = "login"
)

type UserHandler struct {
	emailRegexp    *regexp.Regexp
	passwordRegexp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

type UserClaims struct {
	jwt.RegisteredClaims
	Uid       int64
	Email     string
	Phone     string
	UserAgent string
}

func NewUserHandler(svc service.UserService, codeSvc service.CodeService) UserHandler {
	return UserHandler{
		emailRegexp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRegexp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	ug := server.Group("/users")
	ug.POST("/signup", h.SignUp)
	ug.POST("/login", h.LoginJWT)
	ug.POST("/edit", h.Edit)
	ug.GET("/profile", h.Profile)

	// 手机验证码登录相关功能
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", h.LoginSMS)

}

func (h *UserHandler) SignUp(ctx *gin.Context) {
	type req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var r req
	if err := ctx.Bind(&r); err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	isEmail, err := h.emailRegexp.MatchString(r.Email)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}

	if !isEmail {
		ctx.String(http.StatusOK, "邮箱格式错误")
		return
	}

	err = h.svc.Signup(ctx, domain.User{
		Email:    r.Email,
		Password: r.Password,
	})

	switch err {
	case nil:
		ctx.String(http.StatusOK, "一切正常")
	case service.ErrDuplicateEmail:
		ctx.String(http.StatusOK, "邮箱冲突，请换一个")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}

}
func (h *UserHandler) LoginJWT(ctx *gin.Context) {
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
		h.setJwtToken(ctx, u)
		ctx.String(http.StatusOK, "登陆成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码不对")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) setJwtToken(ctx *gin.Context, u domain.User) {
	uc := UserClaims{
		Uid:       u.Id,
		Email:     u.Email,
		Phone:     u.Phone,
		UserAgent: ctx.GetHeader("User-Agent"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, uc)
	tokenStr, err := token.SignedString(JWTKey)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
	}
	ctx.Header("x-jwt-token", tokenStr)
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

func (h *UserHandler) Edit(ctx *gin.Context) {

}

func (h *UserHandler) Profile(ctx *gin.Context) {
	uc, ok := ctx.MustGet("user").(UserClaims)
	if !ok {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	u, err := h.svc.FindById(ctx, uc.Uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统异常")
		return
	}
	type User struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
		Phone    string `json:"phone"`
		AboutMe  string `json:"aboutMe"`
		Birthday string `json:"birthday"`
	}
	ctx.JSON(http.StatusOK, User{
		Nickname: u.Nickname,
		Email:    u.Email,
		Phone:    u.Phone,
		AboutMe:  u.AboutMe,
		Birthday: u.Birthday.Format(time.DateOnly),
	})
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var r Req
	if err := ctx.Bind(&r); err != nil {
		return
	}
	if r.Phone == "" {
		ctx.JSON(http.StatusOK, Response{
			Code: 4,
			Msg:  "请输入手机号",
		})
		return
	}
	err := h.codeSvc.Send(ctx, bizLogin, r.Phone)
	switch {
	case err == nil:
		ctx.JSON(http.StatusOK, Response{
			Code: 200,
			Msg:  "发送成功",
		})
	case errors.Is(err, service.ErrCodeSendTooMany):
		ctx.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, Response{
			Code: 500,
			Msg:  "系统错误," + err.Error(),
		})
	}
}

func (h *UserHandler) LoginSMS(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var r Req
	if err := ctx.Bind(&r); err != nil {
		return
	}

	ok, err := h.codeSvc.Verify(ctx, bizLogin, r.Phone, r.Code)
	if err != nil {
		ctx.JSON(http.StatusOK, Response{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}
	if !ok {
		ctx.JSON(http.StatusOK, Response{
			Code: 400,
			Msg:  "验证码错误，请重新输入",
		})
		return
	}

	user, err := h.svc.FindOrCreate(ctx, r.Phone)
	if err != nil {
		ctx.JSON(http.StatusOK, Response{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}
	h.setJwtToken(ctx, user)

	ctx.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "登录成功",
	})
}
