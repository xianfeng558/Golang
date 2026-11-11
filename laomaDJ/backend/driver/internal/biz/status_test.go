package biz

import "testing"

func TestCanTransitDriver(t *testing.T) {
	if !CanTransitDriver(DriverStatusStop, DriverStatusIn) {
		t.Fatal("停用应能进入收车")
	}
	if !CanTransitDriver(DriverStatusListen, DriverStatusOut) {
		t.Fatal("听单应能出车")
	}
	if CanTransitDriver(DriverStatusStop, DriverStatusOut) {
		t.Fatal("停用不能直接出车")
	}
}
