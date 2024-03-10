package auth

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
)

var JWTKey = []byte("9dKfy1k348sDkf329skdFjie9120d8fd")

type SMSService struct {
	svc sms.Service
	key []byte
}

func NewSMSService(svc sms.Service) sms.Service {
	return &SMSService{svc: svc, key: JWTKey}
}

type SMSClaims struct {
	jwt.RegisteredClaims
	Tpl string
}

func (s *SMSService) Send(ctx context.Context, tplToken string, args []string, numbers ...string) error {
	if skipAuth, ok := ctx.Value(sms.SkipAuth).(bool); ok && skipAuth {
		return s.svc.Send(ctx, tplToken, args, numbers...)
	}

	var claims SMSClaims
	_, err := jwt.ParseWithClaims(tplToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}

	return s.svc.Send(ctx, claims.Tpl, args, numbers...)
}