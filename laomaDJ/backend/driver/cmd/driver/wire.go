//go:build wireinject
// +build wireinject

// wireinject 这个构建标签让本文件只在执行 wire 时编译。日常 go run 用的是已经生成的 wire_gen.go。

package main

import (
	"driver/internal/biz"
	"driver/internal/conf"
	"driver/internal/data"
	"driver/internal/server"
	"driver/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp 声明依赖关系。第一个参数是模板留下的服务配置，里面有 Consul 和 Jaeger 地址。函数体里的 panic 不会在运行时执行。
func wireApp(*conf.Service, *conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
