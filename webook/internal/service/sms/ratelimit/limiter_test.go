package ratelimit

import (
	"context"
	"errors"
	mocksms "github.com/jayleonc/geektime-go/webook/internal/service/sms/mocks"
	limitermocks "github.com/jayleonc/geektime-go/webook/pkg/limiter/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRateLimitSMSService_Send(t *testing.T) {
	testCases := []struct {
		name           string
		limitResult    bool
		limitError     error
		sendCall       bool
		wantSendError  error
		wantLimitError error
	}{
		{
			name:        "不限流，成功发送",
			limitResult: false,
			sendCall:    true,
		},
		{
			name:           "限流，发送失败",
			limitResult:    true,
			wantLimitError: errLimited,
		},
		{
			name:           "限流器错误",
			limitError:     errors.New("redis限流器错误"),
			wantLimitError: errors.New("redis限流器错误"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocksms.NewMockService(ctrl)
			l := limitermocks.NewMockLimiter(ctrl)

			l.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(tc.limitResult, tc.limitError)
			if tc.sendCall {
				svc.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(tc.wantSendError)
			}

			service := NewRateLimitSMSService(svc, l)
			err := service.Send(context.Background(), "abc", []string{"123"}, "123456")
			assert.Equal(t, tc.wantLimitError, err)

		})
	}
}
