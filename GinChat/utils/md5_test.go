package utils

import "testing"

// TestPasswordRoundTrip 锁住三件事：摘要和明文不同，正确密码能通过，错误密码不能通过。
// 最后一行对一下 md5("ab") 的标准结果，防止编码时漏了字节或用错了大小写。
func TestPasswordRoundTrip(t *testing.T) {
	hashed := MakePassword("pass123", "000001")
	if hashed == "pass123" {
		t.Fatal("密码不应原文保存")
	}
	if !ValidPassword("pass123", "000001", hashed) {
		t.Fatal("相同密码和盐应该通过")
	}
	if ValidPassword("other", "000001", hashed) {
		t.Fatal("错误密码不应该通过")
	}
	if Md5Encode("ab") != "187ef4436122d1cc2f40dc2b92f0eba0" {
		t.Fatalf("md5(ab) = %s", Md5Encode("ab"))
	}
}
