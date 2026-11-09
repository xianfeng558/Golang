package biz

import (
	"context"
	"testing"
)

type stubRules struct {
	rule *PriceRule
}

func (s stubRules) GetRule(uint, int) (*PriceRule, error) {
	return s.rule, nil
}

// TestGetPriceIncludesStartDistance 覆盖起步里程内的行程。
// 3000 米不到 5 公里，120 秒是 2 分钟。300 分起步价加 2 分钟 * 10 分 = 320 分。
func TestGetPriceIncludesStartDistance(t *testing.T) {
	vb := NewValuationBiz(stubRules{rule: &PriceRule{PriceRuleWork: PriceRuleWork{
		StartFee: 300, DistanceFee: 35, DurationFee: 10,
	}}})
	// 3 公里、2 分钟：里程没有超出 5 公里，只收起步价和时长。
	got, err := vb.GetPrice(context.Background(), "3000", "120", 1, 9)
	if err != nil {
		t.Fatal(err)
	}
	if got != 320 {
		t.Fatalf("short trip price = %d, want 320", got)
	}
}

// TestGetPriceExtraKilometers 覆盖超出起步里程。
// 8000 米是 8 公里，超出 3 公里；600 秒是 10 分钟。300 + 35*3 + 10*10 = 505 分。
func TestGetPriceExtraKilometers(t *testing.T) {
	vb := NewValuationBiz(stubRules{rule: &PriceRule{PriceRuleWork: PriceRuleWork{
		StartFee: 300, DistanceFee: 35, DurationFee: 10,
	}}})
	// 8 公里超出 3 公里，10 分钟。300 + 35*3 + 10*10 = 505。
	got, err := vb.GetPrice(context.Background(), "8000", "600", 1, 9)
	if err != nil {
		t.Fatal(err)
	}
	if got != 505 {
		t.Fatalf("long trip price = %d, want 505", got)
	}
}
