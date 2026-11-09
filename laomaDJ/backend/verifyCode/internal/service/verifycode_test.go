package service

import (
	"testing"

	pb "verifyCode/api/verifyCode"
)

// TestRandCodeDigit 要求 6 位并且每一位都是数字。
func TestRandCodeDigit(t *testing.T) {
	code := RandCode(6, pb.TYPE_DIGIT)
	if len(code) != 6 {
		t.Fatalf("len = %d, code = %s", len(code), code)
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non digit %q in %s", c, code)
		}
	}
}

// TestRandCodeRejectsUnknown 确认未定义的类型返回空串，不会被当成数字验证码。
func TestRandCodeRejectsUnknown(t *testing.T) {
	if RandCode(4, pb.TYPE(99)) != "" {
		t.Fatal("unknown type should return empty")
	}
}
