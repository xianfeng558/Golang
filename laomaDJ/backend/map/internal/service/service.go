package service

import "github.com/google/wire"

// ProviderSet 是 service 层交给 wire 的构造函数。驾车距离入口是 NewMapServiceService。
var ProviderSet = wire.NewSet(NewGreeterService, NewMapServiceService)
