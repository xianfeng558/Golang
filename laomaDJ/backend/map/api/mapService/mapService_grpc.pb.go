// 由 protoc-gen-go-grpc 生成。不要手动修改。
// 版本：
// - protoc-gen-go-grpc 版本 v1.2.0
// - protoc 版本 v3.21.6
// 来源：api/mapService/mapService.proto

package mapService

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

// MapServiceClient 是 MapService 服务的客户端接口。
//
// ctx 的用法，以及如何结束流式调用，见 https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MapServiceClient interface {
	GetDrivingInfo(ctx context.Context, in *GetDrivingInfoReq, opts ...grpc.CallOption) (*GetDrivingReply, error)
}

type mapServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMapServiceClient(cc grpc.ClientConnInterface) MapServiceClient {
	return &mapServiceClient{cc}
}

func (c *mapServiceClient) GetDrivingInfo(ctx context.Context, in *GetDrivingInfoReq, opts ...grpc.CallOption) (*GetDrivingReply, error) {
	out := new(GetDrivingReply)
	err := c.cc.Invoke(ctx, "/api.mapService.MapService/GetDrivingInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MapServiceServer 是 MapService 服务的服务端接口。
// 所有实现都必须嵌入 UnimplementedMapServiceServer
// 以便以后增加方法时仍然兼容。
type MapServiceServer interface {
	GetDrivingInfo(context.Context, *GetDrivingInfoReq) (*GetDrivingReply, error)
	mustEmbedUnimplementedMapServiceServer()
}

// 必须嵌入 UnimplementedMapServiceServer，这样以后服务增加方法时旧实现仍能编译。
type UnimplementedMapServiceServer struct {
}

func (UnimplementedMapServiceServer) GetDrivingInfo(context.Context, *GetDrivingInfoReq) (*GetDrivingReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDrivingInfo not implemented")
}
func (UnimplementedMapServiceServer) mustEmbedUnimplementedMapServiceServer() {}

// 嵌入 UnsafeMapServiceServer 可以放弃这项前向兼容。
// 不建议使用这个接口。MapServiceServer 一旦增加方法，
// 会导致编译错误。
type UnsafeMapServiceServer interface {
	mustEmbedUnimplementedMapServiceServer()
}

func RegisterMapServiceServer(s grpc.ServiceRegistrar, srv MapServiceServer) {
	s.RegisterService(&MapService_ServiceDesc, srv)
}

func _MapService_GetDrivingInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDrivingInfoReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MapServiceServer).GetDrivingInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.mapService.MapService/GetDrivingInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MapServiceServer).GetDrivingInfo(ctx, req.(*GetDrivingInfoReq))
	}
	return interceptor(ctx, in, info, handler)
}

// MapService_ServiceDesc 是 MapService 服务的 grpc.ServiceDesc。
// 它只该直接交给 grpc.RegisterService 使用，
// 不要查看它的内部，也不要修改（复制一份也不要改）。
var MapService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.mapService.MapService",
	HandlerType: (*MapServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDrivingInfo",
			Handler:    _MapService_GetDrivingInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/mapService/mapService.proto",
}
