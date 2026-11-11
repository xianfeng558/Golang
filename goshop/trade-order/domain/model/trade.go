// Package model 是订单表 trade_order。类型名 TraderOrder 沿用课程原文。
// OrderStatus：1 待支付，2 已关闭，3 已支付，4 已发货，5 已收货，6 已完成，7 已追评。
package model

import "time"

type TraderOrder struct {
	ID                    int32     `gorm:"primaryKey;comment:订单主键" json:"id"`
	OrderNo               string    `gorm:"comment:订单号" json:"orderNo"`
	UserId                int32     `gorm:"default:1;comment:下单用户 ID" json:"userId"`
	TotalAmount           float32   `gorm:"comment:商品总额，单位元" json:"totalAmount"`
	ShippingAmount        float32   `gorm:"comment:运费，单位元" json:"shippingAmount"`
	DiscountAmount        float32   `gorm:"comment:优惠金额，单位元" json:"discountAmount"`
	PayAmount             float32   `gorm:"comment:应付金额，单位元" json:"payAmount"`
	RefundAmount          float32   `gorm:"comment:退款金额，单位元" json:"refundAmount"`
	SubmitTime            time.Time `gorm:"comment:提交时间" json:"submitTime"`
	ExpireTime            time.Time `gorm:"comment:支付截止时间" json:"expireTime"`
	AutoReceiveTime       time.Time `gorm:"comment:自动收货时间" json:"autoReceiveTime"`
	ReceiveTime           time.Time `gorm:"comment:实际收货时间" json:"receiveTime"`
	AutoPraise            time.Time `gorm:"comment:自动好评时间" json:"autoPraise"`
	AfterSaleDeadlineTime time.Time `gorm:"comment:售后截止时间" json:"afterSaleDeadlineTime"`
	OrderStatus           int32     `gorm:"default:1;comment:1 待支付，2 已关闭，3 已支付，4 已发货，5 已收货，6 已完成，7 已追评" json:"orderStatus"`
	OrderSource           int32     `gorm:"default:6;comment:订单来源" json:"orderSource"`
	CancelReason          string    `gorm:"comment:取消原因" json:"cancelReason"`
	OrderType             int32     `gorm:"default:1;comment:订单类型" json:"orderType"`
	CreateUser            int32     `gorm:"default:1;comment:创建人" json:"createUser"`
	CreateTime            time.Time `gorm:"comment:创建时间" json:"createTime"`
	UpdateUser            int32     `gorm:"comment:更新人" json:"updateUser"`
	UpdateTime            time.Time `gorm:"comment:更新时间" json:"updateTime"`
	IsDeleted             bool      `gorm:"comment:软删除标记，回滚时要显式更新 false" json:"isDeleted"`
	PayType               int32     `gorm:"default:1;comment:支付方式" json:"payType"`
	IsPackageFree         int32     `gorm:"default:1;comment:是否包邮" json:"isPackageFree"`
}

func (TraderOrder) TableComment() string { return "订单头" }

func (table *TraderOrder) TableName() string {
	return "trade_order"
}
