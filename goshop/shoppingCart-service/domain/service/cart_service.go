// Package service 把购物车的增删改查和算价交给 repository。SQL 不写在这一层。
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

func (u *CartService) AddCart(req *proto.AddCartReq) (*model.ShoppingCart, error) {
	return u.cartRepository.AddCart(req)
}

func (u *CartService) UpdateCart(req *proto.UpdateCartReq) (*model.ShoppingCart, error) {
	return u.cartRepository.UpdateCart(req)
}

func (u *CartService) GetOrderTotal(int32List []int32) (float32, error) {
	return u.cartRepository.GetOrderTotal(int32List)
}

func (u *CartService) FindCart(req *proto.FindCartReq) (*model.ShoppingCart, error) {
	return u.cartRepository.FindCart(req)
}
