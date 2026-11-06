package common

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusBoot 在后台打开一个 HTTP 端口，路径 /metrics 给 Prometheus 来拉指标。
// go-micro 的 prometheus 插件会把每个 RPC 的耗时、次数注册进默认注册表，这个页面就能看到。
// 端口被占用时只打日志，不退出进程，避免监控组件把订单服务一起带走。
func PrometheusBoot(port int) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	addr := "127.0.0.1:" + strconv.Itoa(port)
	go func() {
		log.Println("prometheus metrics:", "http://"+addr+"/metrics")
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Println("prometheus listen failed:", err)
		}
	}()
}
