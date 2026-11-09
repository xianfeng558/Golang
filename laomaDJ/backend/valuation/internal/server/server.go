package server

import (
	"github.com/google/wire"
)

// ProviderSet 把 HTTP 8300 和 gRPC 9300 的构造函数交给 wire。询价走 9300。
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
