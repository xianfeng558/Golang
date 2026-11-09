// Package server 组装司机的 HTTP、gRPC，以及登录后的令牌比对。
package server

import (
	"context"
	"strings"

	"driver/internal/service"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// DriverToken 在 JWT 签名通过之后，再和司机表里的令牌比对。
// jti 必须是手机号。表里的令牌为空、或和请求头 Bearer 后面的字符串不同，都返回 401。
// 比对通过后把司机行放进 context，键的类型是 driverContextKeyType，避免和别的包的字符串键冲突。
func DriverToken(svc *service.DriverService) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			claims, ok := jwt.FromContext(ctx)
			if !ok {
				return nil, errors.Unauthorized("UNAUTHORIZED", "未登录")
			}
			claimsMap, ok := claims.(jwtv4.MapClaims)
			if !ok {
				return nil, errors.Unauthorized("UNAUTHORIZED", "令牌格式错误")
			}
			tel, _ := claimsMap["jti"].(string)
			if tel == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "令牌里没有手机号")
			}

			saved, err := svc.Bz.DI.GetToken(ctx, tel)
			if err != nil || saved == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "登录已失效")
			}
			header, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, errors.Unauthorized("UNAUTHORIZED", "缺少请求头")
			}
			parts := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
			if len(parts) != 2 || parts[1] != saved {
				return nil, errors.Unauthorized("UNAUTHORIZED", "令牌已更新，请重新登录")
			}

			driver, err := svc.Bz.DI.FetchInfoByTel(ctx, tel)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", "司机不存在")
			}
			return handler(context.WithValue(ctx, driverContextKey, driver), req)
		}
	}
}

// driverContextKeyType 是上下文键的类型。空结构体不占数据，只用来区分键，防止别的中间件用同名字符串覆盖。
type driverContextKeyType struct{}

// driverContextKey 是「当前登录司机」在 context 里的键。
var driverContextKey = driverContextKeyType{}
