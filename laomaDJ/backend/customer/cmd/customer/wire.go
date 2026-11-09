//go:build wireinject
// +build wireinject

// wireinject 这个构建标签让本文件只在执行 wire 时编译。日常 go run 用的是已经生成的 wire_gen.go。

package main

import (
	"customer/internal/biz"
	"customer/internal/conf"
	"customer/internal/data"
	"customer/internal/server"
	"customer/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp 声明依赖关系。函数体里的 panic 不会在运行时执行，wire 会按 ProviderSet 生成真正的组装代码。
func wireApp(*conf.Server, *conf.Data, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
