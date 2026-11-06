// Package service 转发购物车的增删改查和算价。算价的 SQL 在 repository。
package service

import (
	"shoppingCart/domain/model"
	"shoppingCart/domain/repository"
	"shoppingCart/proto"
)

type ICartService interface {
	AddCart(*proto.AddCartReq) (obj *model.ShoppingCart, err error)
	UpdateCart(*proto.UpdateCartReq) (obj *model.ShoppingCart, err error)
	GetOrderTotal(int32List []int32) (obj float32, err error)
	FindCart(*proto.FindCartReq) (obj *model.ShoppingCart, err error)
}
type CartService struct {
	cartRepository repository.ICartRepository
}

func NewCartService(cartRepository repository.ICartRepository) ICartService {
	return &CartService{cartRepository: cartRepository}
}

// "number": 0,
//
//	"productId": 0,
//	"productSkuId": 0
//
// 重写接口方法
func (u *CartService) AddCart(req *proto.AddCartReq) (obj *model.ShoppingCart, err error) {
	return u.cartRepository.AddCart(req)
}
func (u *CartService) UpdateCart(req *proto.UpdateCartReq) (obj *model.ShoppingCart, err error) {
	return u.cartRepository.UpdateCart(req)
}
func (u *CartService) GetOrderTotal(int32List []int32) (obj float32, err error) {
	return u.cartRepository.GetOrderTotal(int32List)
}
func (u *CartService) FindCart(req *proto.FindCartReq) (obj *model.ShoppingCart, err error) {
	return u.cartRepository.FindCart(req)
}
