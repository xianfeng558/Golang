// 由 Wire 生成。不要手动修改。

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"driver/internal/biz"
	"driver/internal/conf"
	"driver/internal/data"
	"driver/internal/server"
	"driver/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

import (
	_ "go.uber.org/automaxprocs"
)

// 下面的组装函数由 wire.go 生成：

// wireApp 组装并返回 Kratos 应用。
func wireApp(confService *conf.Service, confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	dataData, cleanup, err := data.NewData(confData, confService, logger)
	if err != nil {
		return nil, nil, err
	}
	greeterRepo := data.NewGreeterRepo(dataData, logger)
	greeterUsecase := biz.NewGreeterUsecase(greeterRepo, logger)
	greeterService := service.NewGreeterService(greeterUsecase)
	driverInterface := data.NewDriverInterface(dataData)
	driverBiz := biz.NewDriverBiz(driverInterface)
	driverService := service.NewDriverService(driverBiz)
	grpcServer := server.NewGRPCServer(confServer, greeterService, driverService, logger)
	httpServer := server.NewHTTPServer(confServer, greeterService, driverService, logger)
	app := newApp(confService, logger, grpcServer, httpServer)
	return app, func() {
		cleanup()
	}, nil
}
