// Package repository 把 proto 里的订单转成表记录。
// 订单号形如 Y + 时间 + 用户 ID + 6 位随机数，用来给支付回调对账。
package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"trade-order/domain/model"
	"trade-order/proto"

	"gorm.io/gorm"
)

/**
 int32 clientId = 1;
    string phone = 2;
    int32 systemId = 3;
    string verificationCode = 4;
**/
//接口
type ITradeRepository interface {
	FindOrder(req *proto.FindOrderReq) (*model.TraderOrder, error)
	AddTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error)
	UpdateTradeOrder(req *proto.AddTradeOrderReq) (*model.TraderOrder, error)
}

// 创建实例
func NewTradeRepository(db *gorm.DB) ITradeRepository {
	return &TradeRepository{mysqlDB: db}
}

// 数据DB
type TradeRepository struct {
	mysqlDB *gorm.DB
}

func SwapToStruct(req, target interface{}) (err error) {
	dataByte, err := json.Marshal(req)
	if err != nil {
		return
	}
	err = json.Unmarshal(dataByte, target)
	return err
}

// 新增订单
func (u *TradeRepository) AddTradeOrder(req *proto.AddTradeOrderReq) (obj *model.TraderOrder, err error) {
	trade := &model.TraderOrder{}
	err = SwapToStruct(req.TradeOrder, trade)
	if err != nil {
		log.Println("SwapToStruct  err :", err)
	}
	log.Println("SwapToStruct  trade :", trade)
	now := time.Now()
	trade.CreateTime = now //
	trade.SubmitTime = now
	tp, _ := time.ParseDuration("30m")
	trade.ExpireTime = now.Add(tp) //订单失效时间 30m后
	trade.OrderNo = getOrderNo(now, trade.UserId)

	tb := u.mysqlDB.Create(&trade)
	//err = errors.New("400", "测试异常", 400)
	fmt.Println("repository AddTradeOrder   >>>> ", trade)
	return trade, tb.Error //err
}

// 修改订单
func (u *TradeRepository) UpdateTradeOrder(req *proto.AddTradeOrderReq) (obj *model.TraderOrder, err error) {
	trade := model.TraderOrder{}
	trade.ID = req.TradeOrder.ID
	trade.OrderStatus = req.TradeOrder.OrderStatus
	trade.IsDeleted = req.TradeOrder.IsDeleted
	trade.UpdateTime = time.Now() //
	tb := u.mysqlDB.Model(&model.TraderOrder{}).
		Where("id = ?", trade.ID).
		Updates(&trade)
	//Update("order_status", trade.OrderStatus)
	//err = errors.New("400", "测试异常", 400)
	fmt.Println("repository UpdateTradeOrder   >>>> ", trade)
	return &trade, tb.Error //err
}

func (u *TradeRepository) FindOrder(req *proto.FindOrderReq) (obj *model.TraderOrder, err error) {
	id := req.GetId()
	no := req.GetOrderNo()
	obj = &model.TraderOrder{}
	tb := u.mysqlDB.Where("id=? or order_no=?", id, no).Find(obj)
	fmt.Println("FindTradeOrder>>>>>>> ", obj)
	return obj, tb.Error
}

// 生产 订单号   Y2022 06 27 11 00 53 948 97 103564
//
//	年    月 日 时  分 秒 毫秒 ID  随机数
func getOrderNo(time2 time.Time, userID int32) string {
	var tradeNo string
	tempNum := strconv.Itoa(rand.Intn(999999-100000+1) + 100000)
	tradeNo = "Y" + time2.Format("20060102150405.000") + strconv.Itoa(int(userID)) + tempNum
	tradeNo = strings.Replace(tradeNo, ".", "", -1)
	return tradeNo
}
