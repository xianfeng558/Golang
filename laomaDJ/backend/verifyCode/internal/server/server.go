package server

import (
	"github.com/google/wire"
)

// ProviderSet 把 HTTP 8000 和 gRPC 9000 的构造函数交给 wire。生成验证码走 9000。
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
