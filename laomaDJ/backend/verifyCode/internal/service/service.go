package service

import "github.com/google/wire"

// ProviderSet 是 service 层交给 wire 的构造函数。生成验证码的入口是 NewVerifyCodeService。
var ProviderSet = wire.NewSet(NewGreeterService, NewVerifyCodeService)
