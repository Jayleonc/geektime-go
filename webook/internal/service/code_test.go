package service

import (
	"testing"
)

func TestCodeService_generate(t *testing.T) {
	svc := codeService{}
	t.Log(svc.generate())
}
