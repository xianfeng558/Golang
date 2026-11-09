package biz

import (
	"context"
	"customer/api/valuation"
	"database/sql"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"gorm.io/gorm"
)

const CustomerSecret = "yourSecretKey"
const CustomerDuration = 2 * 30 * 24 * 3600

// 模型
type Customer struct {
	// 业务逻辑
	CustomerWork
	// token部分
	CustomerToken
	// 基础字段
	gorm.Model
}

// 业务逻辑部分
type CustomerWork struct {
	Telephone string         `gorm:"type:varchar(15);uniqueIndex;" json:"telephone"`
	Name      sql.NullString `gorm:"type:varchar(255);uniqueIndex;" json:"name"`
	Email     sql.NullString `gorm:"type:varchar(255);uniqueIndex;" json:"email"`
	Wechat    sql.NullString `gorm:"type:varchar(255);uniqueIndex;" json:"wechat"`
	CityID    uint           `gorm:"index;" json:"city_id"`
}

// token 部分
type CustomerToken struct {
	Token          string       `gorm:"type:varchar(4095);" json:"token"`
	TokenCreatedAt sql.NullTime `gorm:"" json:"token_created_at"`
}

type CustomerBiz struct {
}

func NewCustomerBiz() *CustomerBiz {
	return &CustomerBiz{}
}

func (cb *CustomerBiz) GetEstimatePrice(ctx context.Context, origin, destination string) (int64, error) {
	// 一，grpc 获取
	// 1.获取consul客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "localhost:8500"
	consulClient, err := api.NewClient(consulConfig)
	// 2.获取服务发现管理器
	dis := consul.New(consulClient)
	if err != nil {
		return 0, err
	}
	// 2.1,连接目标grpc服务器
	endpoint := "discovery:///Valuation"

	conn, err := grpc.DialInsecure(
		//context.Background(),
		ctx,
		grpc.WithEndpoint(endpoint), // 目标服务的名字
		// 使用服务发现
		grpc.WithDiscovery(dis),
		// 中间件
		grpc.WithMiddleware(
			// tracing 的客户端中间件
			tracing.Client(),
		),
	)
	if err != nil {
		return 0, nil
	}
	//关闭
	defer func() {
		_ = conn.Close()
	}()

	// 2.2,发送获取验证码请求
	client := valuation.NewValuationClient(conn)
	reply, err := client.GetEstimatePrice(context.Background(), &valuation.GetEstimatePriceReq{
		Origin:      origin,
		Destination: destination,
	})
	if err != nil {
		return 0, err
	}

	return reply.Price, nil
}
