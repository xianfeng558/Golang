// Package handler 实现加购、改购物车、按 ID 查询、按购物车 ID 列表算总价。
package handler

import (
	"context"
	"fmt"
	"shoppingCart/domain/service"
	"shoppingCart/proto"
)

// CartHandler 实现购物车的四个 RPC。字段用接口，方便以后替换存储方式。
type CartHandler struct {
	CartService service.ICartService
}

// AddCart 插入一行购物车。成功时把新行的数量和主键填回响应，调用方靠 ID 做后续修改。
// int64(...) 是类型转换：表里的数量是 int32，协议里这个字段声明成了 int64。
func (u *CartHandler) AddCart(ctx context.Context, req *proto.AddCartReq, resp *proto.AddCartResp) error {
	obj, err := u.CartService.AddCart(req)
	if err != nil {
		println("  AddCart err :", err)
	} else {
		resp.CanSetShoppingCartNumber = int64(obj.Number)
		resp.ShoppingCartNumber = int64(obj.Number)
		resp.IsBeyondMaxLimit = false // 查询sku
		resp.ID = obj.ID              //增加新增cart的ID
		fmt.Println(" AddCart  handler  >>>>>>  ", resp)
	}
	return err
}

// UpdateCart 按购物车 ID 更新数量、删除标记等。订单回滚时会把 IsDeleted 改回 false。
func (u *CartHandler) UpdateCart(ctx context.Context, req *proto.UpdateCartReq, resp *proto.UpdateCartResp) error {
	obj, err := u.CartService.UpdateCart(req)
	if err != nil {
		println("  UpdateCart err :", err)
	} else {
		resp.CanSetShoppingCartNumber = int64(obj.Number)
		resp.ShoppingCartNumber = int64(obj.Number)
		resp.IsBeyondMaxLimit = false // 查询sku
		resp.ID = obj.ID              //增加新增cart的ID
		fmt.Println(" UpdateCart  handler  >>>>>>  ", resp)
	}
	return err
}

// FindCart 按 ID 查一条购物车。这里只回了编号、用户和是否删除，下单前用它判断购物车还在不在。
func (u *CartHandler) FindCart(ctx context.Context, req *proto.FindCartReq, obj *proto.FindCartResp) error {
	//int32List := SplitToInt32List(req.GetCartIds(), ",")
	cart, err := u.CartService.FindCart(req)
	obj.ShoppingCart = &proto.ShoppingCart{}
	obj.ShoppingCart.Id = cart.ID
	obj.ShoppingCart.UserId = cart.UserId
	obj.ShoppingCart.IsDeleted = cart.IsDeleted
	//其他需要再加
	return err
}

// GetOrderTotal 把若干购物车 ID 的「数量 × 规格售价」加总。
// 第二个返回值用 _ 丢掉。算价失败时总价会是 0，调用方目前看不出来，这是课程留下的写法。
func (u *CartHandler) GetOrderTotal(ctx context.Context, req *proto.OrderTotalReq, obj *proto.OrderTotalResp) error {
	//int32List := SplitToInt32List(req.GetCartIds(), ",")
	obj.TotalPrice, _ = u.CartService.GetOrderTotal(req.GetCartIds())
	return nil
}
