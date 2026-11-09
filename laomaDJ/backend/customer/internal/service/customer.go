// Package service 是乘客接口的入口。
// 它不写 SQL。手机号规则在 biz，Redis、MySQL 和验证码 RPC 在 data。
package service

import (
	"context"
	"time"

	pb "customer/api/customer"
	"customer/internal/biz"
	"customer/internal/data"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// CustomerService 是乘客 HTTP 的入口。
// 业务失败时把 code 设为 1，并且第二个返回值是 nil。
// 这样 HTTP 仍是 200，页面可以直接读 JSON 里的 message，不用先处理 4xx。
// 未登录访问受保护接口时，中间件会在进入这里之前返回 HTTP 401。
type CustomerService struct {
	pb.UnimplementedCustomerServer
	CD *data.CustomerData // 验证码、乘客表、JWT 落库
	cb *biz.CustomerBiz   // 询价
}

// NewCustomerService 供 wire 组装。
func NewCustomerService(cd *data.CustomerData, cb *biz.CustomerBiz) *CustomerService {
	return &CustomerService{CD: cd, cb: cb}
}

// GetVerifyCode 对应 GET /customer/get-verify-code/{telephone}。
// 手机号不合法时不调用验证码服务。合法时要 6 位数字，写入 Redis 键 CVC:<手机号>，有效 60 秒。
// 没有短信通道，verifyCode 直接出现在响应里，只适合学习。
// 验证码服务或 Redis 失败时 code 为 1，message 是「验证码获取失败」，不把内部错误原文抛给浏览器。
func (s *CustomerService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	if !biz.ValidTelephone(req.Telephone) {
		return &pb.GetVerifyCodeResp{Code: 1, Message: "电话号码格式错误"}, nil
	}
	code, err := s.CD.FetchVerifyCode(ctx, req.Telephone)
	if err != nil {
		return &pb.GetVerifyCodeResp{Code: 1, Message: "验证码获取失败"}, nil
	}
	return &pb.GetVerifyCodeResp{
		Code:           0,
		Message:        "SUCCESS",
		VerifyCode:     code,
		VerifyCodeTime: time.Now().Unix(),
		VerifyCodeLife: biz.VerifyCodeLife,
	}, nil
}

// Login 对应 POST /customer/login，正文是 JSON。
// 先读 Redis。键不存在、或和请求里的验证码不同，都返回「验证码不匹配」。
// 比对通过后按手机号取乘客，没有就插入一行，再用乘客主键作为 JWT 的 jti 签发令牌。
// 令牌会写回乘客表。同一乘客再次登录会覆盖旧令牌，上一张立即失效。
func (s *CustomerService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	code := s.CD.GetVerifyCode(ctx, req.Telephone)
	if code == "" || code != req.VerifyCode {
		return &pb.LoginResp{Code: 1, Message: "验证码不匹配"}, nil
	}
	customer, err := s.CD.GetCustomerByTelephone(req.Telephone)
	if err != nil {
		return &pb.LoginResp{Code: 1, Message: "顾客信息获取错误"}, nil
	}
	token, err := s.CD.GenerateTokenAndSave(customer, biz.TokenTTL(), biz.CustomerSecret)
	if err != nil {
		return &pb.LoginResp{Code: 1, Message: "Token生成失败"}, nil
	}
	return &pb.LoginResp{
		Code:          0,
		Message:       "login success",
		Token:         token,
		TokenCreateAt: time.Now().Unix(),
		TokenLife:     biz.CustomerDuration,
	}, nil
}

// Logout 对应 GET /customer/logout。调用前中间件已经验过签名，并确认令牌和表里的一致。
// 这里再从 JWT 的 jti 取出乘客主键，把表里的 token 列清掉。
// 清掉之后，这张令牌即使还没到过期时间，也不能再询价。
func (s *CustomerService) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	id, err := customerID(ctx)
	if err != nil {
		return &pb.LogoutResp{Code: 1, Message: err.Error()}, nil
	}
	if err := s.CD.DelToken(id); err != nil {
		return &pb.LogoutResp{Code: 1, Message: "Token删除失败"}, nil
	}
	return &pb.LogoutResp{Code: 0, Message: "logout success"}, nil
}

// EstimatePrice 对应 GET /customer/estimate-price/{origin}/{destination}。
// 必须带 Authorization: Bearer <登录令牌>。起点或终点为空时不调用下游。
// price 单位是分。下游失败时 message 里是计价或地图返回的原因，例如高德错误、数据库错误。
// 高德每次给出的秒数可能不同，所以同一对坐标的金额会变。
func (s *CustomerService) EstimatePrice(ctx context.Context, req *pb.EstimatePriceReq) (*pb.EstimatePriceResp, error) {
	if req.Origin == "" || req.Destination == "" {
		return &pb.EstimatePriceResp{Code: 1, Message: "起点或终点为空"}, nil
	}
	price, err := s.cb.GetEstimatePrice(ctx, req.Origin, req.Destination)
	if err != nil {
		return &pb.EstimatePriceResp{Code: 1, Message: err.Error()}, nil
	}
	return &pb.EstimatePriceResp{
		Code:        0,
		Message:     "SUCCESS",
		Origin:      req.Origin,
		Destination: req.Destination,
		Price:       price,
	}, nil
}

// customerID 从已经通过签名校验的 JWT 里取出 jti。
// 乘客令牌的 jti 是 customers.id 的十进制字符串，不是手机号。
// 断言失败时返回错误，不让类型不对的 claims 把进程打崩。
func customerID(ctx context.Context) (string, error) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return "", errString("未登录")
	}
	claimsMap, ok := claims.(jwtv4.MapClaims)
	if !ok {
		return "", errString("令牌格式错误")
	}
	id, _ := claimsMap["jti"].(string)
	if id == "" {
		return "", errString("令牌里没有乘客 ID")
	}
	return id, nil
}

type errString string

func (e errString) Error() string { return string(e) }
