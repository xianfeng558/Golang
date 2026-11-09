// Package data 是脚手架留下的数据层。地图服务不连 MySQL，也不连 Redis。
// 配置文件里的 database、redis 没有被读取。删掉 NewData 会让 wire_gen.go 编译失败。
package data

import (
	"map/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet 交给 wire。NewGreeterRepo 只为了脚手架，地图业务不用它。
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data 是空的。地图距离来自高德 HTTP，不经过这里。
type Data struct {
}

// NewData 不打开数据库。cleanup 只打一行日志，进程退出时由 Kratos 调用。
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
