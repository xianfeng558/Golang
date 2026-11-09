// Package service 是司机 HTTP 和 gRPC 的入口。业务判断在 biz，存储在 data。
package service

import (
	"context"
	"time"

	pb "driver/api/driver"
	"driver/internal/biz"

	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv4 "github.com/golang-jwt/jwt/v4"
)

// DriverService 把请求交给 DriverBiz，再把结果填进 protobuf 响应。
// 业务失败时 code 为 1，HTTP 仍返回 200，message 里是原因。
type DriverService struct {
	pb.UnimplementedDriverServer
	Bz *biz.DriverBiz
}

// NewDriverService 供 wire 组装。
func NewDriverService(bz *biz.DriverBiz) *DriverService {
	return &DriverService{Bz: bz}
}

// IDNoCheck 对应 POST /driver/idno-check。
// 还没有接身份证核验服务。返回 code 1 和明确说明，避免调用方把空响应当成核验通过。
func (s *DriverService) IDNoCheck(ctx context.Context, req *pb.IDNoCheckReq) (*pb.IDNoCheckResp, error) {
	return &pb.IDNoCheckResp{Code: 1, Message: "身份证校验尚未实现"}, nil
}

// GetVerifyCode 对应 GET /driver/get-verify-code/{telephone}。
// 成功时响应里的 verifyCode 和 Redis 里的值相同，有效 60 秒。没有短信。
func (s *DriverService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	code, err := s.Bz.GetVerifyCode(ctx, req.Telephone)
	if err != nil {
		return &pb.GetVerifyCodeResp{Code: 1, Message: err.Error()}, nil
	}
	return &pb.GetVerifyCodeResp{
		Code:           0,
		Message:        "SUCCESS",
		VerifyCode:     code,
		VerifyCodeTime: time.Now().Unix(),
		VerifyCodeLife: biz.VerifyCodeLife,
	}, nil
}

// SubmitPhone 对应 POST /driver/submit-phone。
// 新手机号会建成 stop。已经存在时返回原来的状态，不因为唯一索引再插入一次。
func (s *DriverService) SubmitPhone(ctx context.Context, req *pb.SubmitPhoneReq) (*pb.SubmitPhoneResp, error) {
	driver, err := s.Bz.InitDriverInfo(ctx, req.Telephone)
	if err != nil {
		return &pb.SubmitPhoneResp{Code: 1, Message: err.Error()}, nil
	}
	return &pb.SubmitPhoneResp{
		Code:    0,
		Message: "司机号码提交成功",
		Status:  driver.Status.String,
	}, nil
}

// Login 对应 POST /driver/login。
// 成功时 token 的 jti 是手机号。tokenLife 是 30 天的秒数。验证码错误时 message 为「验证码不匹配」。
func (s *DriverService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	token, err := s.Bz.CheckLogin(ctx, req.Telephone, req.VerifyCode)
	if err != nil {
		return &pb.LoginResp{Code: 1, Message: err.Error()}, nil
	}
	return &pb.LoginResp{
		Code:          0,
		Message:       "司机登录成功",
		Token:         token,
		TokenCreateAt: time.Now().Unix(),
		TokenLife:     biz.DriverTokenLife,
	}, nil
}

// Logout 对应 DELETE /driver/logout。
// 从令牌的 jti 取出手机号，把司机表的 token 列写成 NULL。需要先通过 JWT 中间件。
func (s *DriverService) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	tel, err := telephoneFromToken(ctx)
	if err != nil {
		return &pb.LogoutResp{Code: 1, Message: err.Error()}, nil
	}
	if err := s.Bz.DI.ClearToken(ctx, tel); err != nil {
		return &pb.LogoutResp{Code: 1, Message: "退出失败"}, nil
	}
	return &pb.LogoutResp{Code: 0, Message: "logout success"}, nil
}

// telephoneFromToken 读取 JWT 的 jti。司机令牌里它是手机号。
// 类型断言用逗号形式，避免 claims 不是 MapClaims 时进程崩溃。
func telephoneFromToken(ctx context.Context) (string, error) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return "", errString("未登录")
	}
	claimsMap, ok := claims.(jwtv4.MapClaims)
	if !ok {
		return "", errString("令牌格式错误")
	}
	tel, _ := claimsMap["jti"].(string)
	if tel == "" {
		return "", errString("令牌里没有手机号")
	}
	return tel, nil
}

type errString string

func (e errString) Error() string { return string(e) }
