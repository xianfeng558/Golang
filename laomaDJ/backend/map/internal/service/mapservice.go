// Package service 把高德结果转成地图服务的 gRPC 响应。失败时用 Internal，让计价服务能读到原因。
package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	pb "map/api/mapService"
	"map/internal/biz"
)

// MapServiceService 把驾车距离和时长从高德转成 gRPC 响应。
// 名字里两个 Service 是因为 proto 服务叫 MapService，这一层又是 Kratos 的 service。
type MapServiceService struct {
	pb.UnimplementedMapServiceServer
	msbiz *biz.MapServiceBiz
}

// NewMapServiceService 供 wire 组装。
func NewMapServiceService(msbiz *biz.MapServiceBiz) *MapServiceService {
	return &MapServiceService{msbiz: msbiz}
}

// GetDrivingInfo 返回距离（米）和时长（秒），都是字符串，和计价服务约定的类型一致。
// 高德失败时错误码是 LBS_ERROR。不要用普通数字当 gRPC 状态码，否则调用方只能看到连接被断开。
func (s *MapServiceService) GetDrivingInfo(ctx context.Context, req *pb.GetDrivingInfoReq) (*pb.GetDrivingReply, error) {
	distance, duration, err := s.msbiz.GetDrivingInfo(ctx, req.Origin, req.Destination)
	if err != nil {
		return nil, errors.InternalServer("LBS_ERROR", err.Error())
	}
	return &pb.GetDrivingReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Distance:    distance,
		Duration:    duration,
	}, nil
}
