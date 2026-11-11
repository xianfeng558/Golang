package model

// 订单状态和课程约定一致，支付回调会把待支付改成已支付。
const (
	OrderWaitPay  int32 = 1 // 待支付
	OrderClosed   int32 = 2 // 已关闭
	OrderPaid     int32 = 3 // 已支付
	OrderShipped  int32 = 4 // 已发货
	OrderReceived int32 = 5 // 已收货
	OrderFinished int32 = 6 // 已完成
	OrderReviewed int32 = 7 // 已追评
)

// CanTransit 判断订单能不能从 from 变成 to。
// 相同状态视为成功，方便支付回调重试。
// 待支付只能关闭或支付。已支付之后按发货、收货、完成、追评往前走，不能从已发货跳回待支付。
func CanTransit(from, to int32) bool {
	if from == to {
		return true
	}
	switch from {
	case 0, OrderWaitPay:
		return to == OrderClosed || to == OrderPaid
	case OrderPaid:
		return to == OrderShipped || to == OrderClosed
	case OrderShipped:
		return to == OrderReceived
	case OrderReceived:
		return to == OrderFinished
	case OrderFinished:
		return to == OrderReviewed
	default:
		return false
	}
}
