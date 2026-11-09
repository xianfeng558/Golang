// Package data 是脚手架留下的数据层。验证码服务不存码，配置里的 MySQL 和 Redis 没有被读取。
package data

import (
	"verifyCode/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet 交给 wire。验证码的生成在 service，不经过这个 Data。
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo)

// Data 是空的。码由调用方自己写入 Redis。
type Data struct {
}

// NewData 不打开数据库。参数 c 保留是因为 wire 生成的构造函数签名不能随便改。
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
