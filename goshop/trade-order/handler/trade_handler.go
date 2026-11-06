// Package handler 是订单 RPC 入口。支付服务改状态时调用 UpdateTradeOrder。
package handler

import (
	"context"
	"fmt"
	common "git.mashibing.com/msb_47094/shopping-comm"
	"trade-order/domain/service"
	"trade-order/proto"
)

// TradeOrderHandler 实现创建订单、改状态、按编号查询。支付回调改的就是这里的状态。
type TradeOrderHandler struct {
	TradeOrderService service.ITradeOrderService
}

// AddTradeOrder 创建订单。订单号、创建时间和 30 分钟后的过期时间在 repository 里补上。
func (u *TradeOrderHandler) AddTradeOrder(ctx context.Context, req *proto.AddTradeOrderReq, resp *proto.AddTradeOrderResp) error {
	obj, err := u.TradeOrderService.AddTradeOrder(req)
	if err != nil {
		println("  AddTradeOrder err :", err)
	} else {
		fmt.Println(obj.UpdateTime)
		fmt.Println(" AddTradeOrder  handler  >>>>>>  ", resp)
	}
	return err
}

// UpdateTradeOrder 只改状态和删除标记。支付成功时支付服务把 OrderStatus 设为 3。
func (u *TradeOrderHandler) UpdateTradeOrder(ctx context.Context, req *proto.AddTradeOrderReq, resp *proto.AddTradeOrderResp) error {
	obj, err := u.TradeOrderService.UpdateTradeOrder(req)
	if err != nil {
		println("  UpdateTradeOrder err :", err)
	} else {
		fmt.Println(obj.UpdateTime)
		fmt.Println(" UpdateTradeOrder  handler  >>>>>>  ", resp)
	}
	return err
}

// FindOrder 按订单主键或订单号查询，再把表记录转成 proto。
// SwapToStruct 靠 json 标签对字段。时间类型对不上字符串时，对应字段会是空的。
func (u *TradeOrderHandler) FindOrder(ctx context.Context, req *proto.FindOrderReq, resp *proto.FindOrderResp) error {
	obj, err := u.TradeOrderService.FindOrder(req)
	if err != nil {
		println("  FindTradeOrder err :", err)
	} else {
		order := &proto.TradeOrder{}
		err := common.SwapToStruct(obj, order)
		if err != nil {
			fmt.Println("转换失败 ", err)
		}
		resp.TradeOrder = order
	}
	return err
}
