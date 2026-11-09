package biz

import "github.com/google/wire"

// ProviderSet 是 biz 层交给 wire 的构造函数。组装结果在 cmd/driver/wire_gen.go，那个文件不要手改。
var ProviderSet = wire.NewSet(NewGreeterUsecase, NewDriverBiz)
