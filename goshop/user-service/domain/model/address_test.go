package model

import "testing"

func TestCheckAddress(t *testing.T) {
	ok := UserAddress{UserID: 2, Receiver: "张三", Phone: "13800000000", Detail: "中关村大街 1 号"}
	if err := CheckAddress(ok); err != nil {
		t.Fatal(err)
	}
	bad := ok
	bad.Phone = "12345"
	if err := CheckAddress(bad); err == nil {
		t.Fatal("短手机号应失败")
	}
}
