package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jayleonc/geektime-go/webook/internal/repository"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms"
	"github.com/jayleonc/geektime-go/webook/internal/service/sms/middleware"
	"math/rand"
)

var ErrCodeSendTooMany = repository.ErrCodeSendTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, inputCode string) (bool, error)
}

type codeService struct {
	repo repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo repository.CodeRepository, sms sms.Service) CodeService {
	return &codeService{
		repo: repo,
		sms:  sms,
	}
}

func (svc *codeService) Send(ctx context.Context, biz, phone string) error {
	code := svc.generate()
	err := svc.repo.Set(ctx, biz, phone, code) // 将 code 存储起来
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	token, err := middleware.GenerateToken(codeTplId)
	if err != nil {
		return err
	}
	return svc.sms.Send(ctx, token, []string{code}, phone) // 将 code 发送出去
}

func (svc *codeService) Verify(ctx context.Context, biz, phone, inputCode string) (bool, error) {
	ok, err := svc.repo.Verify(ctx, biz, phone, inputCode)
	// 对外面屏蔽验证次数过多的错误，直接告诉调用者，验证码错误
	if errors.Is(err, repository.ErrCodeVerifyTooMany) {
		return false, nil
	}
	return ok, err
}

func (svc *codeService) generate() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
