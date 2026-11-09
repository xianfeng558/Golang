package server

import (
	"github.com/google/wire"
)

// ProviderSet 把 HTTP 8200 和 gRPC 9200 的构造函数交给 wire。驾车距离走 9200。
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
