package server

import (
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	v1 "valuation/api/helloworld/v1"
	"valuation/api/valuation"
	"valuation/internal/conf"
	"valuation/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer 监听计价 gRPC，默认 0.0.0.0:9300。乘客询价打的就是这个端口。
// tracing.Server 给每次调用建 span。Jaeger 没开时 span 发不出去，接口结果不受影响。
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService,
	valuationService *service.ValuationService,
	logger log.Logger) *grpc.Server {

	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			// 加入 tracing 中间件
			tracing.Server(),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterGreeterServer(srv, greeter)
	// 注册
	valuation.RegisterValuationServer(srv, valuationService)
	return srv
}
