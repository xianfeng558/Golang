// Package server 组装乘客的 HTTP 和 gRPC，并放登录校验、跨域。
// 跨域在路由匹配之前处理，浏览器从 5173 访问 8100 时会先发 OPTIONS。
// JWT 签名通过之后，还要和 customers.token 里保存的字符串比对，退出后旧令牌失效。
package server

import (
	"context"
	"strconv"

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
	registerTripRoutes(srv, customerService)
	return srv
}

type tripBody struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Price       int64  `json:"price"`
	ID          uint   `json:"id"`
	Status      string `json:"status"`
}

// registerTripRoutes 挂上叫车接口。它们不在 proto 里，所以在这里手写。
// 服务器中间件仍会校验 JWT，和询价使用同一张令牌。
func registerTripRoutes(srv *http.Server, customerService *service.CustomerService) {
	r := srv.Route("/customer")
	r.POST("/trips", func(ctx http.Context) error {
		var body tripBody
		if err := ctx.Bind(&body); err != nil {
			return err
		}
		reply, err := customerService.CreateTrip(ctx, body.Origin, body.Destination, body.Price)
		if err != nil {
			return err
		}
		return ctx.JSON(200, reply)
	})
	r.GET("/trips", func(ctx http.Context) error {
		reply, err := customerService.ListTrips(ctx)
		if err != nil {
			return err
		}
		return ctx.JSON(200, reply)
	})
	r.POST("/trips/move", func(ctx http.Context) error {
		var body tripBody
		if err := ctx.Bind(&body); err != nil {
			return err
		}
		if body.ID == 0 {
			if id, err := strconv.ParseUint(ctx.Query().Get("id"), 10, 64); err == nil {
				body.ID = uint(id)
			}
		}
		reply, err := customerService.MoveTrip(ctx, body.ID, body.Status)
		if err != nil {
			return err
		}
		return ctx.JSON(200, reply)
	})
}
