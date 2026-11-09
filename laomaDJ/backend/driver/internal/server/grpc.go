package server

import (
	"driver/api/driver"
	v1 "driver/api/helloworld/v1"
	"driver/internal/conf"
	"driver/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer 监听司机的 gRPC，默认 0.0.0.0:9400。
// 同时注册了脚手架 Greeter 和司机业务。页面和 curl 走的是 HTTP 8400，这个端口留给别的进程调用。
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService,
	driverService *service.DriverService,
	logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
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
	driver.RegisterDriverServer(srv, driverService)
	return srv
}
