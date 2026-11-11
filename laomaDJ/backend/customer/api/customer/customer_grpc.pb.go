// 由 protoc-gen-go-grpc 生成。不要手动修改。
// 版本：
// - protoc-gen-go-grpc 版本 v1.2.0
// - protoc 版本 v3.21.6
// 来源：api/customer/customer.proto

package customer

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

// CustomerClient 是 Customer 服务的客户端接口。
//
// ctx 的用法，以及如何结束流式调用，见 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CustomerClient interface {
	// 获取验证码
	GetVerifyCode(ctx context.Context, in *GetVerifyCodeReq, opts ...grpc.CallOption) (*GetVerifyCodeResp, error)
	// 登录
	Login(ctx context.Context, in *LoginReq, opts ...grpc.CallOption) (*LoginResp, error)
	// 退出
	Logout(ctx context.Context, in *LogoutReq, opts ...grpc.CallOption) (*LogoutResp, error)
	// 价格预估
	EstimatePrice(ctx context.Context, in *EstimatePriceReq, opts ...grpc.CallOption) (*EstimatePriceResp, error)
}

type customerClient struct {
	cc grpc.ClientConnInterface
}

func NewCustomerClient(cc grpc.ClientConnInterface) CustomerClient {
	return &customerClient{cc}
}

func (c *customerClient) GetVerifyCode(ctx context.Context, in *GetVerifyCodeReq, opts ...grpc.CallOption) (*GetVerifyCodeResp, error) {
	out := new(GetVerifyCodeResp)
	err := c.cc.Invoke(ctx, "/api.customer.Customer/GetVerifyCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customerClient) Login(ctx context.Context, in *LoginReq, opts ...grpc.CallOption) (*LoginResp, error) {
	out := new(LoginResp)
	err := c.cc.Invoke(ctx, "/api.customer.Customer/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customerClient) Logout(ctx context.Context, in *LogoutReq, opts ...grpc.CallOption) (*LogoutResp, error) {
	out := new(LogoutResp)
	err := c.cc.Invoke(ctx, "/api.customer.Customer/Logout", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *customerClient) EstimatePrice(ctx context.Context, in *EstimatePriceReq, opts ...grpc.CallOption) (*EstimatePriceResp, error) {
	out := new(EstimatePriceResp)
	err := c.cc.Invoke(ctx, "/api.customer.Customer/EstimatePrice", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CustomerServer 是 Customer 服务的服务端接口。
// 所有实现都必须嵌入 UnimplementedCustomerServer
// 以便以后增加方法时仍然兼容。
type CustomerServer interface {
	// 获取验证码
	GetVerifyCode(context.Context, *GetVerifyCodeReq) (*GetVerifyCodeResp, error)
	// 登录
	Login(context.Context, *LoginReq) (*LoginResp, error)
	// 退出
	Logout(context.Context, *LogoutReq) (*LogoutResp, error)
	// 价格预估
	EstimatePrice(context.Context, *EstimatePriceReq) (*EstimatePriceResp, error)
	mustEmbedUnimplementedCustomerServer()
}

// 必须嵌入 UnimplementedCustomerServer，这样以后服务增加方法时旧实现仍能编译。
type UnimplementedCustomerServer struct {
}

func (UnimplementedCustomerServer) GetVerifyCode(context.Context, *GetVerifyCodeReq) (*GetVerifyCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVerifyCode not implemented")
}
func (UnimplementedCustomerServer) Login(context.Context, *LoginReq) (*LoginResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedCustomerServer) Logout(context.Context, *LogoutReq) (*LogoutResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedCustomerServer) EstimatePrice(context.Context, *EstimatePriceReq) (*EstimatePriceResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EstimatePrice not implemented")
}
func (UnimplementedCustomerServer) mustEmbedUnimplementedCustomerServer() {}

// 嵌入 UnsafeCustomerServer 可以放弃这项前向兼容。
// 不建议使用这个接口。CustomerServer 一旦增加方法，
// 会导致编译错误。
type UnsafeCustomerServer interface {
	mustEmbedUnimplementedCustomerServer()
}

func RegisterCustomerServer(s grpc.ServiceRegistrar, srv CustomerServer) {
	s.RegisterService(&Customer_ServiceDesc, srv)
}

func _Customer_GetVerifyCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVerifyCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomerServer).GetVerifyCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.customer.Customer/GetVerifyCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomerServer).GetVerifyCode(ctx, req.(*GetVerifyCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customer_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomerServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.customer.Customer/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomerServer).Login(ctx, req.(*LoginReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customer_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogoutReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomerServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.customer.Customer/Logout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomerServer).Logout(ctx, req.(*LogoutReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Customer_EstimatePrice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EstimatePriceReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CustomerServer).EstimatePrice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.customer.Customer/EstimatePrice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CustomerServer).EstimatePrice(ctx, req.(*EstimatePriceReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Customer_ServiceDesc 是 Customer 服务的 grpc.ServiceDesc。
// 它只该直接交给 grpc.RegisterService 使用，
// 不要查看它的内部，也不要修改（复制一份也不要改）。
var Customer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.customer.Customer",
	HandlerType: (*CustomerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetVerifyCode",
			Handler:    _Customer_GetVerifyCode_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _Customer_Login_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _Customer_Logout_Handler,
		},
		{
			MethodName: "EstimatePrice",
			Handler:    _Customer_EstimatePrice_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/customer/customer.proto",
}
