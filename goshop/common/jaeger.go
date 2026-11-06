package common

import (
	"io"
	"log"
	"time"

	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

// NewTracer 创建一个 Jaeger 追踪器。
// addr 是 Jaeger Agent 的 UDP 地址，课程里固定用 6831 端口。
// Agent 没启动时，创建通常仍会成功，只是链路数据发不出去，服务本身可以继续跑。
func NewTracer(serviceName string, addr string) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			// const 表示每条请求都采样。学习阶段方便在 Jaeger UI 里看到完整调用链。
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			BufferFlushInterval: 1 * time.Second,
			LogSpans:            true,
			LocalAgentHostPort:  addr,
		},
	}
	return cfg.NewTracer()
}

// UseTracer 创建追踪器并设成全局对象。失败时返回 nil，调用方不要 Close。
func UseTracer(serviceName string) io.Closer {
	tracer, closer, err := NewTracer(serviceName, ConsulIp+":6831")
	if err != nil {
		log.Println("jaeger tracer disabled:", err)
		return nil
	}
	opentracing.SetGlobalTracer(tracer)
	log.Println("jaeger tracer ready, agent", ConsulIp+":6831", "service", serviceName)
	return closer
}
