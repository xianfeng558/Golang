package model

import "time"

// TradeOrderItem 是订单行。一张订单头对应多行，一行记住它来自哪条购物车。
// 购物车和商品在另一个库，交易服务不跨库 JOIN，所以这里只存购物车 ID 和当时的金额快照。
type TradeOrderItem struct {
	ID         int32     `gorm:"primaryKey;comment:订单行主键" json:"id"`
	OrderID    int32     `gorm:"index;comment:订单头 ID" json:"orderId"`
	CartID     int32     `gorm:"index;comment:来源购物车 ID，0 表示没有对应购物车" json:"cartId"`
	Title      string    `gorm:"comment:下单时的商品名快照" json:"title"`
	Quantity   int32     `gorm:"comment:数量" json:"quantity"`
	PayAmount  float32   `gorm:"comment:这一行分摊的应付金额，单位元" json:"payAmount"`
	CreateTime time.Time `gorm:"comment:写入时间" json:"createTime"`
}

func (TradeOrderItem) TableName() string    { return "trade_order_item" }
func (TradeOrderItem) TableComment() string { return "订单行，记下当时的金额快照" }

// TradeOrderLog 记下状态每次变化。面试时用它说明「订单状态不要只改一列，要留流水」。
type TradeOrderLog struct {
	ID         int32     `gorm:"primaryKey;comment:流水主键" json:"id"`
	OrderID    int32     `gorm:"index;comment:订单头 ID" json:"orderId"`
	FromStatus int32     `gorm:"comment:变化前的状态，新建时是 0" json:"fromStatus"`
	ToStatus   int32     `gorm:"comment:变化后的状态" json:"toStatus"`
	Remark     string    `gorm:"comment:原因，例如创建、支付回调" json:"remark"`
	CreateTime time.Time `gorm:"comment:写入时间" json:"createTime"`
}

func (TradeOrderLog) TableName() string    { return "trade_order_log" }
func (TradeOrderLog) TableComment() string { return "订单状态流水" }
