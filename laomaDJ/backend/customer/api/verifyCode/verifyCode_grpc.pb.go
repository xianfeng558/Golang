// 由 protoc-gen-go-grpc 生成。不要手动修改。
// 版本：
// - protoc-gen-go-grpc 版本 v1.2.0
// - protoc 版本 v3.21.6
// 来源：api/verifyCode/verifyCode.proto

package verifyCode

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

// VerifyCodeClient 是 VerifyCode 服务的客户端接口。
//
// ctx 的用法，以及如何结束流式调用，见 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VerifyCodeClient interface {
	GetVerifyCode(ctx context.Context, in *GetVerifyCodeRequest, opts ...grpc.CallOption) (*GetVerifyCodeReply, error)
}

type verifyCodeClient struct {
	cc grpc.ClientConnInterface
}

func NewVerifyCodeClient(cc grpc.ClientConnInterface) VerifyCodeClient {
	return &verifyCodeClient{cc}
}

func (c *verifyCodeClient) GetVerifyCode(ctx context.Context, in *GetVerifyCodeRequest, opts ...grpc.CallOption) (*GetVerifyCodeReply, error) {
	out := new(GetVerifyCodeReply)
	err := c.cc.Invoke(ctx, "/api.verifyCode.VerifyCode/GetVerifyCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VerifyCodeServer 是 VerifyCode 服务的服务端接口。
// 所有实现都必须嵌入 UnimplementedVerifyCodeServer
// 以便以后增加方法时仍然兼容。
type VerifyCodeServer interface {
	GetVerifyCode(context.Context, *GetVerifyCodeRequest) (*GetVerifyCodeReply, error)
	mustEmbedUnimplementedVerifyCodeServer()
}

// 必须嵌入 UnimplementedVerifyCodeServer，这样以后服务增加方法时旧实现仍能编译。
type UnimplementedVerifyCodeServer struct {
}

func (UnimplementedVerifyCodeServer) GetVerifyCode(context.Context, *GetVerifyCodeRequest) (*GetVerifyCodeReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVerifyCode not implemented")
}
func (UnimplementedVerifyCodeServer) mustEmbedUnimplementedVerifyCodeServer() {}

// 嵌入 UnsafeVerifyCodeServer 可以放弃这项前向兼容。
// 不建议使用这个接口。VerifyCodeServer 一旦增加方法，
// 会导致编译错误。
type UnsafeVerifyCodeServer interface {
	mustEmbedUnimplementedVerifyCodeServer()
}

func RegisterVerifyCodeServer(s grpc.ServiceRegistrar, srv VerifyCodeServer) {
	s.RegisterService(&VerifyCode_ServiceDesc, srv)
}

func _VerifyCode_GetVerifyCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVerifyCodeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VerifyCodeServer).GetVerifyCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.verifyCode.VerifyCode/GetVerifyCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VerifyCodeServer).GetVerifyCode(ctx, req.(*GetVerifyCodeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// VerifyCode_ServiceDesc 是 VerifyCode 服务的 grpc.ServiceDesc。
// 它只该直接交给 grpc.RegisterService 使用，
// 不要查看它的内部，也不要修改（复制一份也不要改）。
var VerifyCode_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.verifyCode.VerifyCode",
	HandlerType: (*VerifyCodeServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVerifyCode",
			Handler:    _VerifyCode_GetVerifyCode_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/verifyCode/verifyCode.proto",
}
