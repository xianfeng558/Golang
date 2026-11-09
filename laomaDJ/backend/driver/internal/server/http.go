package server

import (
	"context"
	"driver/api/driver"
	v1 "driver/api/helloworld/v1"
	"driver/internal/biz"
	"driver/internal/conf"
	"driver/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, driverService *service.DriverService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			// JWT 中间件
			selector.Server(
				jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return []byte(biz.SecretKey), nil
				}),
				DriverToken(driverService)).Match(func(ctx context.Context, operation string) bool {
				// 记录不需要的校验的
				noJWT := map[string]struct{}{
					"/api.driver.Driver/Login":         {},
					"/api.driver.Driver/GetVerifyCode": {},
					"/api.driver.Driver/SubmitPhone":   {},
				}
				if _, exists := noJWT[operation]; exists {
					return false
				}
				return true
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)

	driver.RegisterDriverHTTPServer(srv, driverService)
	return srv
}
