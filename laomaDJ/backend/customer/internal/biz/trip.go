package biz

import (
	"time"

	"gorm.io/gorm"
)

// 行程状态。询价成功后乘客确认叫车，才插入一行，状态从待接单开始。
const (
	TripWaitDriver = "wait_driver" // 待接单
	TripOngoing    = "ongoing"     // 行程中
	TripFinished   = "finished"    // 已完成
	TripCanceled   = "canceled"    // 已取消
)

// TripOrder 对应表 trip_orders。金额单位是分，和询价接口一致。
type TripOrder struct {
	ID          uint           `gorm:"primarykey"`
	CreatedAt   time.Time      `gorm:"comment:创建时间"`
	UpdatedAt   time.Time      `gorm:"comment:更新时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:软删除时间，有值时普通查询看不到"`
	CustomerID  uint           `gorm:"index;comment:乘客主键" json:"customerId"`
	Origin      string         `gorm:"comment:起点，经度在前" json:"origin"`
	Destination string         `gorm:"comment:终点，经度在前" json:"destination"`
	Price       int64          `gorm:"comment:预估价格，单位分" json:"price"`
	Status      string         `gorm:"type:varchar(32);index;comment:wait_driver、ongoing、finished、canceled" json:"status"`
}

func (TripOrder) TableComment() string {
	return "行程。询价不写这张表，确认叫车才插入"
}

func (TripOrder) TableName() string { return "trip_orders" }

// CanTransitTrip 规定行程只能往前走，或在结束前取消。已完成不能再取消。
func CanTransitTrip(from, to string) bool {
	switch from + ">" + to {
	case TripWaitDriver + ">" + TripOngoing,
		TripOngoing + ">" + TripFinished,
		TripWaitDriver + ">" + TripCanceled,
		TripOngoing + ">" + TripCanceled:
		return true
	default:
		return false
	}
}
