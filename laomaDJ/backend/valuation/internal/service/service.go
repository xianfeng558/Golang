package service

import "github.com/google/wire"

// ProviderSet 是 service 层交给 wire 的构造函数。询价入口是 NewValuationService。
var ProviderSet = wire.NewSet(NewGreeterService, NewValuationService)
