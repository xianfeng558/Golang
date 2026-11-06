// Package repository 把 proto 里的订单转成表记录。
// 订单号形如 Y + 时间 + 用户 ID + 6 位随机数，用来给支付回调对账。
package repository

import (
	"errors"
	"math/rand"
	"strconv"
	"strings"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	"gorm.io/gorm"
	"trade-order/domain/model"
	"trade-order/proto"
)

// ITradeRepository 是订单的查询、创建和改状态。
type ITradeRepository interface {
	FindOrder(req *proto.FindOrderReq) (*model.TraderOrder, error)
	AddTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error)
	UpdateTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error)
}

func NewTradeRepository(db *gorm.DB) ITradeRepository {
	return &TradeRepository{mysqlDB: db}
}

type TradeRepository struct {
	mysqlDB *gorm.DB
}

// AddTradeOrder 先把请求抄成表结构，再补上时间和订单号。类型名 TraderOrder 沿用课程原文。
func (u *TradeRepository) AddTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error) {
	if req == nil || req.TradeOrder == nil {
		return nil, errors.New("订单内容为空")
	}
	trade := &model.TraderOrder{}
	if err := common.SwapToStruct(req.TradeOrder, trade); err != nil {
		return nil, err
	}
	now := time.Now()
	trade.CreateTime = now
	trade.SubmitTime = now
	trade.ExpireTime = now.Add(30 * time.Minute)
	trade.OrderNo = getOrderNo(now, trade.UserId)
	tb := u.mysqlDB.Create(trade)
	return trade, tb.Error
}

// UpdateTradeOrder 只更新状态、删除标记和更新时间，避免把没传来的字段写成零值。
func (u *TradeRepository) UpdateTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error) {
	if req == nil || req.TradeOrder == nil {
		return nil, errors.New("订单内容为空")
	}
	trade := model.TraderOrder{
		ID:          req.TradeOrder.ID,
		OrderStatus: req.TradeOrder.OrderStatus,
		IsDeleted:   req.TradeOrder.IsDeleted,
		UpdateTime:  time.Now(),
	}
	tb := u.mysqlDB.Model(&model.TraderOrder{}).Where("id = ?", trade.ID).Updates(&trade)
	return &trade, tb.Error
}

func (u *TradeRepository) FindOrder(req *proto.FindOrderReq) (*model.TraderOrder, error) {
	obj := &model.TraderOrder{}
	tb := u.mysqlDB.Where("id=? or order_no=?", req.GetId(), req.GetOrderNo()).Find(obj)
	return obj, tb.Error
}

// getOrderNo 拼出订单号：Y、年月日时分秒毫秒、用户 ID、6 位随机数。
func getOrderNo(time2 time.Time, userID int32) string {
	tempNum := strconv.Itoa(rand.Intn(999999-100000+1) + 100000)
	tradeNo := "Y" + time2.Format("20060102150405.000") + strconv.Itoa(int(userID)) + tempNum
	return strings.ReplaceAll(tradeNo, ".", "")
}
