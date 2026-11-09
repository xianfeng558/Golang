package data

import (
	"valuation/internal/biz"
)

type PriceRuleData struct {
	data *Data
}

func NewPriceRuleInterface(data *Data) biz.PriceRuleInterface {
	return &PriceRuleData{data: data}
}

// PriceRuleData 实现 PriceRuleInterface
func (prd *PriceRuleData) GetRule(cityid uint, curr int) (*biz.PriceRule, error) {
	pr := &biz.PriceRule{}
	result := prd.data.Mdb.Where("city_id=? AND start_at <= ? AND end_at > ?", cityid, curr, curr).First(pr)
	if result.Error != nil {
		return nil, result.Error
	}
	return pr, nil
}
