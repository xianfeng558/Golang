// 本文件在 JWT 签名通过之后，把请求头里的令牌和乘客表里的 token 列比对。

package server

import (
	"context"
	"strings"

	"customer/internal/service"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// customerJWT 确认这张令牌仍然是该乘客表里保存的最新一张。
// 签名校验由前面的 jwt.Server 完成，这里不再验签名。
// 步骤是：从上下文取出 jti（乘客主键），读出表里的 token，再和请求头里 Bearer 后面的字符串逐字比较。
// 对不上说明已经重新登录或退出。claims 类型不对时返回 401，不发生类型断言崩溃。
func customerJWT(customerService *service.CustomerService) middleware.Middleware {
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
			id, _ := claimsMap["jti"].(string)
			if id == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "令牌里没有乘客 ID")
			}

			saved, err := customerService.CD.GetToken(id)
			if err != nil || saved == "" {
				return nil, errors.Unauthorized("UNAUTHORIZED", "登录已失效")
			}

			header, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, errors.Unauthorized("UNAUTHORIZED", "缺少请求头")
			}
			parts := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
			if len(parts) != 2 || parts[1] == "" || parts[1] != saved {
				return nil, errors.Unauthorized("UNAUTHORIZED", "令牌已更新，请重新登录")
			}
			return handler(ctx, req)
		}
	}
}
