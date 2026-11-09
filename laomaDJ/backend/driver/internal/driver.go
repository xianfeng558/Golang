package service

import (
	"context"

	pb "driver/api/driver"
)

type DriverService struct {
	pb.UnimplementedDriverServer
}

func NewDriverService() *DriverService {
	return &DriverService{}
}

func (s *DriverService) IDNoCheck(ctx context.Context, req *pb.IDNoCheckReq) (*pb.IDNoCheckResp, error) {
	return &pb.IDNoCheckResp{}, nil
}
func (s *DriverService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	return &pb.GetVerifyCodeResp{}, nil
}
func (s *DriverService) SubmitPhone(ctx context.Context, req *pb.SubmitPhoneReq) (*pb.SubmitPhoneResp, error) {
	return &pb.SubmitPhoneResp{}, nil
}
func (s *DriverService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	return &pb.LoginResp{}, nil
}
func (s *DriverService) Logout(ctx context.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
	return &pb.LogoutResp{}, nil
}
