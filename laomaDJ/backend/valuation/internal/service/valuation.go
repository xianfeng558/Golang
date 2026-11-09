// Package service 是计价 gRPC 的入口。它先取路程，再用当前小时计价。
package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"valuation/internal/biz"

	pb "valuation/api/valuation"
)

// ValuationService 被乘客服务调用。没有 HTTP 路径。
// 城市写死为 1，因为示例规则只准备了这一座城市。小时用本机时间，所以白天和夜间会命中不同规则。
type ValuationService struct {
	pb.UnimplementedValuationServer
	vb *biz.ValuationBiz
}

// NewValuationService 供 wire 组装。
func NewValuationService(vb *biz.ValuationBiz) *ValuationService {
	return &ValuationService{vb: vb}
}

// GetEstimatePrice 是乘客服务调用的 gRPC 方法。
// 先问地图要米和秒，再用 time.Now().Hour() 选规则。
// 失败时用 Internal 错误把原因传回去。以前用数字 200 当 gRPC 状态码，调用方只能看到 EOF，看不到真正的原因。
func (s *ValuationService) GetEstimatePrice(ctx context.Context, req *pb.GetEstimatePriceReq) (*pb.GetEstimatePriceReply, error) {
	distance, duration, err := s.vb.GetDrivingInfo(ctx, req.Origin, req.Destination)
	if err != nil {
		return nil, errors.InternalServer("MAP_ERROR", err.Error())
	}
	price, err := s.vb.GetPrice(ctx, distance, duration, 1, time.Now().Hour())
	if err != nil {
		return nil, errors.InternalServer("PRICE_ERROR", err.Error())
	}
	return &pb.GetEstimatePriceReply{
		Origin:      req.Origin,
		Destination: req.Destination,
		Price:       price,
	}, nil
}
