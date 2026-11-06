// Package handler 实现加购、改购物车、按 ID 查询、按购物车 ID 列表算总价。
package handler

import (
	"context"

	"shoppingCart/domain/service"
	"shoppingCart/proto"
)

// CartHandler 实现购物车的四个 RPC。字段用接口，方便以后替换存储方式。
type CartHandler struct {
	CartService service.ICartService
}

// AddCart 插入一行购物车。成功时把数量和主键填回响应，调用方靠 ID 做后续修改。
// int64(...) 是类型转换：表里的数量是 int32，协议里这个字段是 int64。
func (u *CartHandler) AddCart(ctx context.Context, req *proto.AddCartReq, resp *proto.AddCartResp) error {
	obj, err := u.CartService.AddCart(req)
	if err != nil {
		return err
	}
	resp.CanSetShoppingCartNumber = int64(obj.Number)
	resp.ShoppingCartNumber = int64(obj.Number)
	resp.IsBeyondMaxLimit = false
	resp.ID = obj.ID
	return nil
}

// UpdateCart 按购物车 ID 更新数量和删除标记。订单回滚时会把 IsDeleted 改回 false。
func (u *CartHandler) UpdateCart(ctx context.Context, req *proto.UpdateCartReq, resp *proto.UpdateCartResp) error {
	obj, err := u.CartService.UpdateCart(req)
	if err != nil {
		return err
	}
	resp.CanSetShoppingCartNumber = int64(obj.Number)
	resp.ShoppingCartNumber = int64(obj.Number)
	resp.IsBeyondMaxLimit = false
	resp.ID = obj.ID
	return nil
}

// FindCart 按 ID 查一条购物车。下单前用它判断购物车还在不在。
func (u *CartHandler) FindCart(ctx context.Context, req *proto.FindCartReq, obj *proto.FindCartResp) error {
	cart, err := u.CartService.FindCart(req)
	if err != nil {
		return err
	}
	obj.ShoppingCart = &proto.ShoppingCart{
		Id:        cart.ID,
		UserId:    cart.UserId,
		IsDeleted: cart.IsDeleted,
	}
	return nil
}

// GetOrderTotal 把若干购物车的「数量 × 规格售价」加总。算价失败时把错误交回去。
func (u *CartHandler) GetOrderTotal(ctx context.Context, req *proto.OrderTotalReq, obj *proto.OrderTotalResp) error {
	total, err := u.CartService.GetOrderTotal(req.GetCartIds())
	if err != nil {
		return err
	}
	obj.TotalPrice = total
	return nil
}
