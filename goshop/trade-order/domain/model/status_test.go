package model

import "testing"

func TestCanTransitPaymentAndClose(t *testing.T) {
	if !CanTransit(OrderWaitPay, OrderPaid) {
		t.Fatal("待支付应能变成已支付")
	}
	if !CanTransit(OrderWaitPay, OrderClosed) {
		t.Fatal("待支付应能关闭")
	}
	if CanTransit(OrderShipped, OrderWaitPay) {
		t.Fatal("已发货不能退回待支付")
	}
	if !CanTransit(OrderPaid, OrderPaid) {
		t.Fatal("相同状态应允许重试")
	}
}
