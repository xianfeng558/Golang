package biz

import "testing"

func TestCanTransitTrip(t *testing.T) {
	if !CanTransitTrip(TripWaitDriver, TripOngoing) {
		t.Fatal("待接单应能变成行程中")
	}
	if !CanTransitTrip(TripOngoing, TripCanceled) {
		t.Fatal("行程中应能取消")
	}
	if CanTransitTrip(TripFinished, TripCanceled) {
		t.Fatal("已完成不能取消")
	}
	if CanTransitTrip(TripWaitDriver, TripFinished) {
		t.Fatal("不能从待接单直接完成")
	}
}
