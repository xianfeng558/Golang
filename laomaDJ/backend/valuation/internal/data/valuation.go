package data

import (
	"strings"

	"valuation/internal/biz"
)

type PriceRuleData struct {
	data *Data
}

func NewPriceRuleInterface(data *Data) biz.PriceRuleInterface {
	return &PriceRuleData{data: data}
}

// GetRule 按城市和当前小时取一条规则。小时区间是左闭右开，SQL 是 start_at <= hour AND end_at > hour。
// 本机 MySQL 偶尔会把空闲连接掐掉。第一次错误文本里包含 invalid connection 时，再查一次。
// 第二次仍失败就把错误返回，不再循环重试。
func (prd *PriceRuleData) GetRule(cityID uint, hour int) (*biz.PriceRule, error) {
	pr, err := prd.one(cityID, hour)
	if err != nil && strings.Contains(err.Error(), "invalid connection") {
		return prd.one(cityID, hour)
	}
	return pr, err
}

// one 执行一次按城市和小时的查询。调用方负责决定要不要重试。
func (prd *PriceRuleData) one(cityID uint, hour int) (*biz.PriceRule, error) {
	pr := &biz.PriceRule{}
	err := prd.data.Mdb.Where("city_id = ? AND start_at <= ? AND end_at > ?", cityID, hour, hour).First(pr).Error
	if err != nil {
		return nil, err
	}
	return pr, nil
}
