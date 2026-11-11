// 由 protoc-gen-go-grpc 生成。不要手动修改。
// 版本：
// - protoc-gen-go-grpc 版本 v1.2.0
// - protoc 版本 v3.21.6
// 来源：api/driver/driver.proto

package driver

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

// DriverClient 是 Driver 服务的客户端接口。
//
// ctx 的用法，以及如何结束流式调用，见 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DriverClient interface {
	// 校验身份证号码
	IDNoCheck(ctx context.Context, in *IDNoCheckReq, opts ...grpc.CallOption) (*IDNoCheckResp, error)
	// 获取验证码
	GetVerifyCode(ctx context.Context, in *GetVerifyCodeReq, opts ...grpc.CallOption) (*GetVerifyCodeResp, error)
	// 提交电话号码
	SubmitPhone(ctx context.Context, in *SubmitPhoneReq, opts ...grpc.CallOption) (*SubmitPhoneResp, error)
	// 登录
	Login(ctx context.Context, in *LoginReq, opts ...grpc.CallOption) (*LoginResp, error)
	// 退出
	Logout(ctx context.Context, in *LogoutReq, opts ...grpc.CallOption) (*LogoutResp, error)
}

type driverClient struct {
	cc grpc.ClientConnInterface
}

func NewDriverClient(cc grpc.ClientConnInterface) DriverClient {
	return &driverClient{cc}
}

func (c *driverClient) IDNoCheck(ctx context.Context, in *IDNoCheckReq, opts ...grpc.CallOption) (*IDNoCheckResp, error) {
	out := new(IDNoCheckResp)
	err := c.cc.Invoke(ctx, "/api.driver.Driver/IDNoCheck", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *driverClient) GetVerifyCode(ctx context.Context, in *GetVerifyCodeReq, opts ...grpc.CallOption) (*GetVerifyCodeResp, error) {
	out := new(GetVerifyCodeResp)
	err := c.cc.Invoke(ctx, "/api.driver.Driver/GetVerifyCode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *driverClient) SubmitPhone(ctx context.Context, in *SubmitPhoneReq, opts ...grpc.CallOption) (*SubmitPhoneResp, error) {
	out := new(SubmitPhoneResp)
	err := c.cc.Invoke(ctx, "/api.driver.Driver/SubmitPhone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *driverClient) Login(ctx context.Context, in *LoginReq, opts ...grpc.CallOption) (*LoginResp, error) {
	out := new(LoginResp)
	err := c.cc.Invoke(ctx, "/api.driver.Driver/Login", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *driverClient) Logout(ctx context.Context, in *LogoutReq, opts ...grpc.CallOption) (*LogoutResp, error) {
	out := new(LogoutResp)
	err := c.cc.Invoke(ctx, "/api.driver.Driver/Logout", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DriverServer 是 Driver 服务的服务端接口。
// 所有实现都必须嵌入 UnimplementedDriverServer
// 以便以后增加方法时仍然兼容。
type DriverServer interface {
	// 校验身份证号码
	IDNoCheck(context.Context, *IDNoCheckReq) (*IDNoCheckResp, error)
	// 获取验证码
	GetVerifyCode(context.Context, *GetVerifyCodeReq) (*GetVerifyCodeResp, error)
	// 提交电话号码
	SubmitPhone(context.Context, *SubmitPhoneReq) (*SubmitPhoneResp, error)
	// 登录
	Login(context.Context, *LoginReq) (*LoginResp, error)
	// 退出
	Logout(context.Context, *LogoutReq) (*LogoutResp, error)
	mustEmbedUnimplementedDriverServer()
}

// 必须嵌入 UnimplementedDriverServer，这样以后服务增加方法时旧实现仍能编译。
type UnimplementedDriverServer struct {
}

func (UnimplementedDriverServer) IDNoCheck(context.Context, *IDNoCheckReq) (*IDNoCheckResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IDNoCheck not implemented")
}
func (UnimplementedDriverServer) GetVerifyCode(context.Context, *GetVerifyCodeReq) (*GetVerifyCodeResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetVerifyCode not implemented")
}
func (UnimplementedDriverServer) SubmitPhone(context.Context, *SubmitPhoneReq) (*SubmitPhoneResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitPhone not implemented")
}
func (UnimplementedDriverServer) Login(context.Context, *LoginReq) (*LoginResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedDriverServer) Logout(context.Context, *LogoutReq) (*LogoutResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Logout not implemented")
}
func (UnimplementedDriverServer) mustEmbedUnimplementedDriverServer() {}

// 嵌入 UnsafeDriverServer 可以放弃这项前向兼容。
// 不建议使用这个接口。DriverServer 一旦增加方法，
// 会导致编译错误。
type UnsafeDriverServer interface {
	mustEmbedUnimplementedDriverServer()
}

func RegisterDriverServer(s grpc.ServiceRegistrar, srv DriverServer) {
	s.RegisterService(&Driver_ServiceDesc, srv)
}

func _Driver_IDNoCheck_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IDNoCheckReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DriverServer).IDNoCheck(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.driver.Driver/IDNoCheck",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DriverServer).IDNoCheck(ctx, req.(*IDNoCheckReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Driver_GetVerifyCode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVerifyCodeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DriverServer).GetVerifyCode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.driver.Driver/GetVerifyCode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DriverServer).GetVerifyCode(ctx, req.(*GetVerifyCodeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Driver_SubmitPhone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SubmitPhoneReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DriverServer).SubmitPhone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.driver.Driver/SubmitPhone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DriverServer).SubmitPhone(ctx, req.(*SubmitPhoneReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Driver_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DriverServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.driver.Driver/Login",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DriverServer).Login(ctx, req.(*LoginReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _Driver_Logout_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LogoutReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DriverServer).Logout(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.driver.Driver/Logout",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DriverServer).Logout(ctx, req.(*LogoutReq))
	}
	return interceptor(ctx, in, info, handler)
}

// Driver_ServiceDesc 是 Driver 服务的 grpc.ServiceDesc。
// 它只该直接交给 grpc.RegisterService 使用，
// 不要查看它的内部，也不要修改（复制一份也不要改）。
var Driver_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.driver.Driver",
	HandlerType: (*DriverServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "IDNoCheck",
			Handler:    _Driver_IDNoCheck_Handler,
		},
		{
			MethodName: "GetVerifyCode",
			Handler:    _Driver_GetVerifyCode_Handler,
		},
		{
			MethodName: "SubmitPhone",
			Handler:    _Driver_SubmitPhone_Handler,
		},
		{
			MethodName: "Login",
			Handler:    _Driver_Login_Handler,
		},
		{
			MethodName: "Logout",
			Handler:    _Driver_Logout_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/driver/driver.proto",
}
