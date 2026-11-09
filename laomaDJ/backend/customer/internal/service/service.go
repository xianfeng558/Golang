package service

import "github.com/google/wire"

// ProviderSet 是 service 层交给 wire 的构造函数。NewGreeterService 是脚手架，代驾接口在 NewCustomerService。
var ProviderSet = wire.NewSet(NewCustomerService, NewGreeterService)
