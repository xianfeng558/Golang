// Package model 是订单表 trade_order。类型名 TraderOrder 沿用课程原文。
// OrderStatus：1 待支付，2 已关闭，3 已支付，4 已发货，5 已收货，6 已完成，7 已追评。
package model

import "time"

type TraderOrder struct {
	//gorm.Model
	ID                    int32     `json:"id"`
	OrderNo               string    `json:"orderNo"`
	UserId                int32     `gorm:"default:1" json:"userId"`
	TotalAmount           float32   `gorm:"total_amount" json:"totalAmount"`
	ShippingAmount        float32   `gorm:"shipping_amount" json:"shippingAmount"`
	DiscountAmount        float32   `gorm:"discount_amount" json:"discountAmount"`
	PayAmount             float32   `gorm:"pay_amount" json:"payAmount"`
	RefundAmount          float32   `gorm:"refund_amount" json:"refundAmount"`
	SubmitTime            time.Time `json:"submitTime"`
	ExpireTime            time.Time `json:"expireTime"`
	AutoReceiveTime       time.Time `json:"autoReceiveTime"`
	ReceiveTime           time.Time `json:"receiveTime"`
	AutoPraise            time.Time `json:"autoPraise"`
	AfterSaleDeadlineTime time.Time `json:"afterSaleDeadlineTime"`
	OrderStatus           int32     `gorm:"default:1" json:"orderStatus"`
	OrderSource           int32     `gorm:"default:6" json:"orderSource"`
	CancelReason          string    `gorm:"cancel_reason" json:"cancelReason"`
	OrderType             int32     `gorm:"default:1" json:"orderType"`
	CreateUser            int32     `gorm:"default:1" json:"createUser"`
	CreateTime            time.Time `json:"createTime"`
	UpdateUser            int32     `json:"updateUser"`
	UpdateTime            time.Time `json:"updateTime"`
	IsDeleted             bool      `json:"isDeleted"`
	PayType               int32     `gorm:"default:1" json:"payType"`
	IsPackageFree         int32     `gorm:"default:1" json:"isPackageFree"`
}

func (table *TraderOrder) TableName() string {
	return "trade_order"
}
