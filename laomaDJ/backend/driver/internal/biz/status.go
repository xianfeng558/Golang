package biz

// CanTransitDriver 规定司机状态怎么变。
// 新登记是 stop。审核通过后进入 in（收车）。听单 listen 和出车 out 可以来回，停用 stop 可以从任意工作状态进入。
func CanTransitDriver(from, to string) bool {
	if from == to {
		return true
	}
	switch from + ">" + to {
	case DriverStatusStop + ">" + DriverStatusIn,
		DriverStatusIn + ">" + DriverStatusListen,
		DriverStatusListen + ">" + DriverStatusOut,
		DriverStatusOut + ">" + DriverStatusListen,
		DriverStatusListen + ">" + DriverStatusIn,
		DriverStatusIn + ">" + DriverStatusStop,
		DriverStatusListen + ">" + DriverStatusStop,
		DriverStatusOut + ">" + DriverStatusStop:
		return true
	default:
		return false
	}
}
