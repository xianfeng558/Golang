package biz

import "testing"

// TestValidTelephone 检查大陆手机号：1 开头，第二位 3-9，共 11 位。短号和空串不通过。
func TestValidTelephone(t *testing.T) {
	if !ValidTelephone("13800000000") {
		t.Fatal("13800000000 should pass")
	}
	if ValidTelephone("12345") || ValidTelephone("") {
		t.Fatal("short number should fail")
	}
}
