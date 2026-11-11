// Package biz 放司机的表结构、状态和登录规则。
// 具体的 SQL、Redis 和验证码 RPC 通过 DriverInterface 交给 data，biz 不 import 数据库驱动。
package biz

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

const (
	// SecretKey 是司机 JWT 的 HS256 密钥，和乘客的 CustomerSecret 不是同一串，两边令牌不能混用。
	SecretKey = "driver secret key"
	// DriverTokenLife 是令牌声明的有效秒数，30 天。表里的 token 被清空后，未过期的旧令牌也不能再用。
	DriverTokenLife = 30 * 24 * 3600
	// VerifyCodeLife 是验证码在 Redis 里的存活秒数。
	VerifyCodeLife int32 = 60

	DriverStatusOut    = "out"    // 出车
	DriverStatusIn     = "in"     // 收车
	DriverStatusListen = "listen" // 听单
	DriverStatusStop   = "stop"   // 停用，新登记的司机先处于这个状态
)

var telephonePattern = regexp.MustCompile(`^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`)

// Driver 对应表 drivers。主键和三个时间列代替 gorm.Model，这样列注释能写进 MySQL。
type Driver struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time      `gorm:"comment:创建时间"`
	UpdatedAt time.Time      `gorm:"comment:更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index;comment:软删除时间，有值时普通查询看不到"`
	DriverWork
}

func (Driver) TableComment() string {
	return "司机。验证码在 Redis 2 号库，键是 DVC 加手机号"
}

// DriverWork 是司机资料。
// 手机号唯一。身份证号也是唯一索引，但当前没有写入身份证的接口，新司机这一列是 NULL。
// Token 可空。退出时写成 NULL，中间件看到空令牌就拒绝访问。
// Status 用枚举列，只允许 out、in、listen、stop 四个值。
type DriverWork struct {
	Telephone     string         `gorm:"type:varchar(16);uniqueIndex;comment:登录手机号，唯一" json:"telephone"`
	Token         sql.NullString `gorm:"type:varchar(2047);comment:当前 JWT，退出后为 NULL" json:"token"`
	Status        sql.NullString `gorm:"type:enum('out', 'in', 'listen', 'stop');comment:out 出车，in 收车，listen 听单，stop 停用" json:"status"`
	Name          sql.NullString `gorm:"type:varchar(255);index;comment:姓名" json:"name"`
	IdNumber      sql.NullString `gorm:"type:char(18);uniqueIndex;comment:身份证号，唯一" json:"id_number"`
	IdImageA      sql.NullString `gorm:"type:varchar(255);comment:身份证人像面地址" json:"id_image_a"`
	LicenseImageA sql.NullString `gorm:"type:varchar(255);comment:驾驶证正本地址" json:"license_image_a"`
	LicenseImageB sql.NullString `gorm:"type:varchar(255);comment:驾驶证副本地址" json:"license_image_b"`
	DistinctCode  sql.NullString `gorm:"type:varchar(16);index;comment:行政区划代码" json:"distinct_code"`
	TelephoneBak  sql.NullString `gorm:"type:varchar(16);index;comment:备用手机号" json:"telephone_bak"`
	AuditAt       sql.NullTime   `gorm:"index;comment:审核通过时间" json:"audit_at"`
}

// DriverInterface 是 data 层必须实现的操作。
// biz 只依赖这个接口，测试时可以换成内存实现。
type DriverInterface interface {
	GetVerifyCode(context.Context, string) (string, error)      // 向验证码服务要码并写入 Redis
	GetSavedVerifyCode(context.Context, string) (string, error) // 读出刚才保存的码
	SaveToken(context.Context, string, string) error            // 按手机号保存 JWT，没有行就先建
	ClearToken(context.Context, string) error                   // 退出时清空令牌
	GetToken(context.Context, string) (string, error)           // 按手机号读当前令牌
	InitDriverInfo(context.Context, string) (*Driver, error)    // 登记为 stop，已存在则返回原行
	FetchInfoByTel(context.Context, string) (*Driver, error)    // 按手机号读整行，给中间件放进上下文
}

// DriverBiz 处理司机登录和注册。
type DriverBiz struct {
	DI DriverInterface
}

// NewDriverBiz 供 wire 组装。
func NewDriverBiz(di DriverInterface) *DriverBiz {
	return &DriverBiz{DI: di}
}

// GetVerifyCode 先校验手机号，再交给 data 去要码。
// 号码不合法时返回「电话号码格式错误」，不会访问验证码服务。
func (db *DriverBiz) GetVerifyCode(ctx context.Context, tel string) (string, error) {
	if !telephonePattern.MatchString(tel) {
		return "", fmt.Errorf("电话号码格式错误")
	}
	return db.DI.GetVerifyCode(ctx, tel)
}

// CheckLogin 比对 Redis 里的验证码，签发 JWT，并把令牌写回司机表。
// 验证码过期、为空、或和请求不一致，都返回错误，不签发令牌。
// 这台手机号还没有司机行时，SaveToken 会先建一条状态为 stop 的记录。
// 返回的字符串是完整 JWT，jti 是手机号。
func (db *DriverBiz) CheckLogin(ctx context.Context, tel, verifyCode string) (string, error) {
	code, err := db.DI.GetSavedVerifyCode(ctx, tel)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(verifyCode) == "" || verifyCode != code {
		return "", fmt.Errorf("验证码不匹配")
	}
	token, err := generateJWT(tel)
	if err != nil {
		return "", err
	}
	if err := db.DI.SaveToken(ctx, tel, token); err != nil {
		return "", err
	}
	return token, nil
}

// InitDriverInfo 登记手机号。号码不合法时不写库。
// 已登记过的手机号由 data 返回原记录，状态保持原值，不会被再次写成 stop。
func (db *DriverBiz) InitDriverInfo(ctx context.Context, tel string) (*Driver, error) {
	if !telephonePattern.MatchString(tel) {
		return nil, fmt.Errorf("电话号码格式错误")
	}
	return db.DI.InitDriverInfo(ctx, tel)
}

// generateJWT 用手机号作为 jti 签发 HS256 令牌。
// 乘客令牌的 jti 是数字主键，司机这里是手机号，中间件据此回表查询。
func generateJWT(tel string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "LaomaDJ",
		Subject:   "driver authentication",
		Audience:  []string{"driver"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(DriverTokenLife * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        tel,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(SecretKey))
}
