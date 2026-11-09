// Package biz 放乘客的业务规则和表结构。
// 这里不直接碰 Redis 和 MySQL。询价要调用别的进程，所以连计价服务的代码也放在这里。
package biz

import (
	"context"
	"customer/api/valuation"
	"customer/internal/discover"
	"database/sql"
	"regexp"
	"time"

	"gorm.io/gorm"
)

const (
	// CustomerSecret 是乘客 JWT 的 HS256 签名密钥。
	// 签发和校验必须用同一串。它写在源码里，只给本机学习，不能保护真实账号。
	CustomerSecret = "yourSecretKey"
	// CustomerDuration 是令牌声明的有效秒数，60 天。
	// 这只是 JWT 自己的过期时间。乘客表里的 token 列被清空后，未过期的旧令牌也不能再用。
	CustomerDuration = 60 * 24 * 3600
	// VerifyCodeLife 是验证码在 Redis 里的存活秒数。超时后键消失，登录会得到「验证码不匹配」。
	VerifyCodeLife = 60
)

// telephonePattern 匹配 11 位中国大陆手机号。
// 在包加载时编译一次。每个请求再编译会重复做同样的工作。
var telephonePattern = regexp.MustCompile(`^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`)

// ValidTelephone 判断 telephone 是不是 11 位大陆手机号。
// 空串、座机、位数不对都返回 false。它不访问数据库。
func ValidTelephone(telephone string) bool {
	return telephonePattern.MatchString(telephone)
}

// Customer 对应表 customers。三个匿名结构体的字段会平铺成同一张表的列，不是三张表。
// gorm.Model 提供 id、created_at、updated_at、deleted_at。deleted_at 有值的行不会被普通查询查到。
type Customer struct {
	CustomerWork  // 手机号、姓名这些业务列
	CustomerToken // 最近一次登录留下的令牌
	gorm.Model
}

// CustomerWork 是乘客资料。
// 姓名、邮箱、微信用 sql.NullString：没有填写时写入 NULL。
// 如果用普通 string，没填就是空串，第二个乘客会撞上唯一索引。
// 唯一列用 varchar(191)，因为 utf8mb4 一个字符最多 4 字节，191×4 不超过旧版 InnoDB 的 1000 字节索引上限。
// varchar(255) 会在建表时返回 Error 1071。
type CustomerWork struct {
	Telephone string         `gorm:"type:varchar(15);uniqueIndex;" json:"telephone"` // 登录账号，11 位手机号
	Name      sql.NullString `gorm:"type:varchar(191);uniqueIndex;" json:"name"`     // 姓名，可空
	Email     sql.NullString `gorm:"type:varchar(191);uniqueIndex;" json:"email"`    // 邮箱，可空
	Wechat    sql.NullString `gorm:"type:varchar(191);uniqueIndex;" json:"wechat"`   // 微信号，可空
	CityID    uint           `gorm:"index;" json:"city_id"`                          // 所在城市。询价目前固定用城市 1，这列先留着
}

// CustomerToken 保存这名乘客当前有效的那一张 JWT。
// 重新登录会覆盖它，退出会把它清成空串。中间件用请求里的令牌和这一列比对。
type CustomerToken struct {
	Token          string       `gorm:"type:varchar(512);" json:"token"` // 完整 JWT。HS256 的三段字符串用 512 足够
	TokenCreatedAt sql.NullTime `json:"token_created_at"`                // 最近一次签发时间。退出后 Valid 为 false，列变成 NULL
}

// CustomerBiz 处理需要调用其它服务的乘客业务。目前只有询价。
type CustomerBiz struct{}

// NewCustomerBiz 供 wire 组装。
func NewCustomerBiz() *CustomerBiz {
	return &CustomerBiz{}
}

// GetEstimatePrice 向计价服务询价。
// origin、destination 都是「经度,纬度」。返回值单位是分，例如 1415 表示 14.15 元。
// Consul 的 8500 通着时按服务名 Valuation 发现，否则直连 127.0.0.1:9300。
// 连接按服务名缓存。计价进程重启后，要连同乘客进程一起重启，否则可能拿到已经断开的连接。
// 出错时价格为 0，err 是拨号失败或计价服务返回的 gRPC 错误。
func (cb *CustomerBiz) GetEstimatePrice(ctx context.Context, origin, destination string) (int64, error) {
	conn, err := discover.Default.Conn("Valuation", "127.0.0.1:9300")
	if err != nil {
		return 0, err
	}
	reply, err := valuation.NewValuationClient(conn).GetEstimatePrice(ctx, &valuation.GetEstimatePriceReq{
		Origin:      origin,
		Destination: destination,
	})
	if err != nil {
		return 0, err
	}
	return reply.GetPrice(), nil
}

// TokenTTL 把 CustomerDuration 换成 time.Duration。
// JWT 的 ExpiresAt 要的是时间长度，调用处不必再写「秒数乘 time.Second」。
func TokenTTL() time.Duration {
	return time.Duration(CustomerDuration) * time.Second
}
