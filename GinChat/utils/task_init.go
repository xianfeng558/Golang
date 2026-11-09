package utils

import "time"

// TimerFunc 是定时器每次要调用的函数。
// 参数 param 原样传进去，当前清理任务用不到，所以 main 传 nil。
// 返回 false 表示这次做完就停；清理心跳的函数一直返回 true。
type TimerFunc func(interface{}) bool

// Timer 在单独的协程里先等 delay，之后每 tick 执行一次 fun。
// delay 对应配置 timeout.DelayHeartbeat，tick 对应 timeout.HeartbeatHz，单位在调用处换成秒。
// 协程自己循环，main 可以继续去听 HTTP 端口。
func Timer(delay, tick time.Duration, fun TimerFunc, param interface{}) {
	go func() {
		if fun == nil {
			return
		}
		timer := time.NewTimer(delay)
		defer timer.Stop()
		for range timer.C {
			if !fun(param) {
				return
			}
			timer.Reset(tick)
		}
	}()
}
