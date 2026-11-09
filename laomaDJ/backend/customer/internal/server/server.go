package server

import (
	"github.com/google/wire"
)

// ProviderSet 把 HTTP 8100 和 gRPC 9100 的构造函数交给 wire。
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
