package service

import (
	"context"

	v1 "customer/api/helloworld/v1"
	"customer/internal/biz"
)

// GreeterService 是脚手架的问候接口。登录、询价不会走到这里。
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc *biz.GreeterUsecase
}

// NewGreeterService 供 wire 组装。保留它，生成的 wire_gen.go 才能编译。
func NewGreeterService(uc *biz.GreeterUsecase) *GreeterService {
	return &GreeterService{uc: uc}
}

// SayHello 把请求里的名字原样放进回复，没有代驾含义。
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	g, err := s.uc.CreateGreeter(ctx, &biz.Greeter{Hello: in.Name})
	if err != nil {
		return nil, err
	}
	return &v1.HelloReply{Message: "Hello " + g.Hello}, nil
}
