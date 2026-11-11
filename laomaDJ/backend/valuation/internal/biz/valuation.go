// Package biz 按地图返回的米和秒，以及城市时段规则，算出预估费用。
package biz

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"valuation/api/mapService"
	"valuation/internal/discover"

	"gorm.io/gorm"
)

// includedKilometers 是起步价里已经包含的公里数。超出的整公里才乘 DistanceFee。
const includedKilometers int64 = 5

// PriceRule 对应表 price_rules。费用字段的单位都是分，不是元。
type PriceRule struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"comment:创建时间"`
	UpdatedAt time.Time      `gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间，有值时普通查询看不到"`
	PriceRuleWork
}

func (PriceRule) TableComment() string { return "计价规则，金额单位是分" }

// PriceRuleWork 是一条城市时段规则。
// StartAt、EndAt 是小时，区间左闭右开。7 和 23 表示从 7 点到 23 点之前，23 点整不再使用这条规则。
type PriceRuleWork struct {
	CityID      uint  `gorm:"comment:城市编号" json:"city_id"`
	StartFee    int64 `gorm:"comment:起步价，单位分，含 5 公里" json:"start_fee"`
	DistanceFee int64 `gorm:"comment:超出起步里程后每公里费用，单位分" json:"distance_fee"`
	DurationFee int64 `gorm:"comment:每分钟费用，单位分" json:"duration_fee"`
	StartAt     int   `gorm:"type:int;comment:开始小时，包含" json:"start_at"`
	EndAt       int   `gorm:"type:int;comment:结束小时，不包含" json:"end_at"`
}

// PriceRuleInterface 按城市和当前小时取出唯一一条规则。查不到时返回错误。
type PriceRuleInterface interface {
	GetRule(cityID uint, hour int) (*PriceRule, error)
}

// ValuationBiz 先问地图要距离和时长，再按规则算价。
type ValuationBiz struct {
	pri PriceRuleInterface
}

// NewValuationBiz 供 wire 组装。
func NewValuationBiz(pri PriceRuleInterface) *ValuationBiz {
	return &ValuationBiz{pri: pri}
}

// GetPrice 把高德给的米和秒换成费用。
// distance、duration 是十进制整数字符串，单位分别是米和秒。cityID 当前由调用方固定传 1。hour 是 0 到 23。
// 公里和分钟都向下取整，不足 1 公里或 1 分钟的零头不计费。
// 总里程不到 includedKilometers 时，里程加价按 0，不会因为减法得到负数。
// 返回值单位是分。解析失败或没有匹配规则时返回错误。
func (vb *ValuationBiz) GetPrice(ctx context.Context, distance, duration string, cityID uint, hour int) (int64, error) {
	rule, err := vb.pri.GetRule(cityID, hour)
	if err != nil {
		return 0, err
	}
	meters, err := strconv.ParseInt(distance, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("距离不是整数: %w", err)
	}
	seconds, err := strconv.ParseInt(duration, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("时长不是整数: %w", err)
	}
	extraKm := meters/1000 - includedKilometers
	if extraKm < 0 {
		extraKm = 0
	}
	minutes := seconds / 60
	return rule.StartFee + rule.DistanceFee*extraKm + rule.DurationFee*minutes, nil
}

// GetDrivingInfo 向地图服务要驾车距离和时长。
// 返回的两个字符串单位是米和秒，还没有换成公里和分钟。
// Consul 可用时按名字 Map 发现，否则直连 127.0.0.1:9200。
// 地图没有返回这两个字段时，不当成 0 公里，而是返回错误，避免用空串算出起步价。
func (vb *ValuationBiz) GetDrivingInfo(ctx context.Context, origin, destination string) (string, string, error) {
	conn, err := discover.Default.Conn("Map", "127.0.0.1:9200")
	if err != nil {
		return "", "", err
	}
	reply, err := mapService.NewMapServiceClient(conn).GetDrivingInfo(ctx, &mapService.GetDrivingInfoReq{
		Origin:      origin,
		Destination: destination,
	})
	if err != nil {
		return "", "", err
	}
	if reply.GetDistance() == "" || reply.GetDuration() == "" {
		return "", "", fmt.Errorf("地图服务没有返回距离或时长")
	}
	return reply.Distance, reply.Duration, nil
}
