package server

import (
	"github.com/google/wire"
)

// ProviderSet 把 HTTP 8400 和 gRPC 9400 的构造函数交给 wire。
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
