// 由 protoc-gen-go-grpc 生成。不要手动修改。
// 版本：
// - protoc-gen-go-grpc 版本 v1.2.0
// - protoc 版本 v3.21.6
// 来源：api/valuation/valuation.proto

package valuation

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// 这是编译期断言，用来保证这份生成文件
// 和当前参与编译的 grpc 包兼容。
// 需要 gRPC-Go v1.32.0 或更高版本。
const _ = grpc.SupportPackageIsVersion7

// ValuationClient 是 Valuation 服务的客户端接口。
//
// ctx 的用法，以及如何结束流式调用，见 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ValuationClient interface {
	GetEstimatePrice(ctx context.Context, in *GetEstimatePriceReq, opts ...grpc.CallOption) (*GetEstimatePriceReply, error)
}

type valuationClient struct {
	cc grpc.ClientConnInterface
}

func NewValuationClient(cc grpc.ClientConnInterface) ValuationClient {
	return &valuationClient{cc}
}

func (c *valuationClient) GetEstimatePrice(ctx context.Context, in *GetEstimatePriceReq, opts ...grpc.CallOption) (*GetEstimatePriceReply, error) {
	out := new(GetEstimatePriceReply)
	err := c.cc.Invoke(ctx, "/api.valuation.Valuation/GetEstimatePrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ValuationServer 是 Valuation 服务的服务端接口。
// 所有实现都必须嵌入 UnimplementedValuationServer
// 以便以后增加方法时仍然兼容。
type ValuationServer interface {
	GetEstimatePrice(context.Context, *GetEstimatePriceReq) (*GetEstimatePriceReply, error)
	mustEmbedUnimplementedValuationServer()
}

// 必须嵌入 UnimplementedValuationServer，这样以后服务增加方法时旧实现仍能编译。
type UnimplementedValuationServer struct {
}

func (UnimplementedValuationServer) GetEstimatePrice(context.Context, *GetEstimatePriceReq) (*GetEstimatePriceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetEstimatePrice not implemented")
}
func (UnimplementedValuationServer) mustEmbedUnimplementedValuationServer() {}

// 嵌入 UnsafeValuationServer 可以放弃这项前向兼容。
// 不建议使用这个接口。ValuationServer 一旦增加方法，
// 会导致编译错误。
type UnsafeValuationServer interface {
	mustEmbedUnimplementedValuationServer()
}

func RegisterValuationServer(s grpc.ServiceRegistrar, srv ValuationServer) {
	s.RegisterService(&Valuation_ServiceDesc, srv)
}

func _Valuation_GetEstimatePrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetEstimatePriceReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ValuationServer).GetEstimatePrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.valuation.Valuation/GetEstimatePrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ValuationServer).GetEstimatePrice(ctx, req.(*GetEstimatePriceReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Valuation_ServiceDesc 是 Valuation 服务的 grpc.ServiceDesc。
// 它只该直接交给 grpc.RegisterService 使用，
// 不要查看它的内部，也不要修改（复制一份也不要改）。
var Valuation_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.valuation.Valuation",
	HandlerType: (*ValuationServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetEstimatePrice",
			Handler:    _Valuation_GetEstimatePrice_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/valuation/valuation.proto",
}
