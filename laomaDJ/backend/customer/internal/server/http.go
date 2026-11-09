// Package server 组装乘客的 HTTP 和 gRPC，并放登录校验、跨域。
// 跨域在路由匹配之前处理，浏览器从 5173 访问 8100 时会先发 OPTIONS。
// JWT 签名通过之后，还要和 customers.token 里保存的字符串比对，退出后旧令牌失效。
package server

import (
	"context"

	"customer/api/customer"
	v1 "customer/api/helloworld/v1"
	"customer/internal/biz"
	"customer/internal/conf"
	"customer/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// NewHTTPServer 监听配置里的 HTTP 地址，默认 0.0.0.0:8100。
// 先经过跨域 Filter，再经过恢复中间件和 JWT。
// 登录、获取验证码不校验令牌。询价和退出要带 Authorization: Bearer。
// operation 用的是 protobuf 的全名，不是 URL。例如 /api.customer.Customer/Login。
// Greeter 是模板接口，和代驾无关，仍注册着，避免生成代码和 wire 对不上。
func NewHTTPServer(c *conf.Server, customerService *service.CustomerService, greeter *service.GreeterService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Filter(Filter()),
		http.Middleware(
			recovery.Recovery(),
			selector.Server(
				jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
					return []byte(biz.CustomerSecret), nil
				}),
				customerJWT(customerService),
			).Match(func(ctx context.Context, operation string) bool {
				noJWT := map[string]struct{}{
					"/api.customer.Customer/Login":         {},
					"/api.customer.Customer/GetVerifyCode": {},
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
	// 注册customer的http服务
	customer.RegisterCustomerHTTPServer(srv, customerService)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}
