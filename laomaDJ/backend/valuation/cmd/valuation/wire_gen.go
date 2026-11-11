// 由 Wire 生成。不要手动修改。

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"valuation/internal/biz"
	"valuation/internal/conf"
	"valuation/internal/data"
	"valuation/internal/server"
	"valuation/internal/service"
)

import (
	_ "go.uber.org/automaxprocs"
)

// 下面的组装函数由 wire.go 生成：

// wireApp 组装并返回 Kratos 应用。
func wireApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	dataData, cleanup, err := data.NewData(confData, logger)
	if err != nil {
		return nil, nil, err
	}
	greeterRepo := data.NewGreeterRepo(dataData, logger)
	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
	greeterService := service.NewGreeterService(greeterUsecase)
	priceRuleInterface := data.NewPriceRuleInterface(dataData)
	valuationBiz := biz.NewValuationBiz(priceRuleInterface)
	valuationService := service.NewValuationService(valuationBiz)
	grpcServer := server.NewGRPCServer(confServer, greeterService, valuationService, logger)
	httpServer := server.NewHTTPServer(confServer, greeterService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
