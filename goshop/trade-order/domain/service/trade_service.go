// Package service 转发创建订单、改订单状态、按 ID 或订单号查询。
package service

import (
	"trade-order/domain/model"
	"trade-order/domain/repository"
	"trade-order/proto"
)

type ITradeOrderService interface {
	FindOrder(req *proto.FindOrderReq) (*model.TraderOrder, error)
	AddTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error)
	UpdateTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error)
}

type TradeOrderService struct {
	tradeRepository repository.ITradeRepository
}

func NewTradeOrderService(tradeRepository repository.ITradeRepository) ITradeOrderService {
	return &TradeOrderService{tradeRepository: tradeRepository}
}

func (u *TradeOrderService) AddTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error) {
	return u.tradeRepository.AddTradeOrder(req)
}

func (u *TradeOrderService) UpdateTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error) {
	return u.tradeRepository.UpdateTradeOrder(req)
}

func (u *TradeOrderService) FindOrder(req *proto.FindOrderReq) (*model.TraderOrder, error) {
	return u.tradeRepository.FindOrder(req)
}
