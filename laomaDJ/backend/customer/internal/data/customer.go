// Package data 负责乘客服务的存储和对外的验证码 RPC。
// 验证码键是 CVC:<手机号>，写在 Redis 1 号库，60 秒过期。
// 表结构类型来自 biz，避免再定义一套模型。
package data

import (
	"context"
	verifyCode "customer/api/verfiyCode"
	"customer/internal/biz"
	"customer/internal/discover"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

// CustomerData 持有连接。方法都通过 data.Rdb 和 data.Mdb 操作，不自己再建客户端。
type CustomerData struct {
	data *Data
}

// NewCustomerData 供 wire 调用。data 由 NewData 创建，里面已经连上 Redis 1 号库和 laomadj_customer。
func NewCustomerData(data *Data) *CustomerData {
	return &CustomerData{data: data}
}

// FetchVerifyCode 向验证码服务要 6 位数字，再按手机号写入 Redis。
// 请求的 Type 为 1，对应验证码服务的 DIGIT。键是 CVC:<手机号>，过期时间是 biz.VerifyCodeLife 秒。
// 同一个手机号再次获取会覆盖旧码。Consul 可用时按名字 VerifyCode 发现，否则直连 127.0.0.1:9000。
// 验证码服务返回空串、或 Redis 写入失败时，不把空码交给调用方。
func (cd *CustomerData) FetchVerifyCode(ctx context.Context, telephone string) (string, error) {
	conn, err := discover.Default.Conn("VerifyCode", "127.0.0.1:9000")
	if err != nil {
		return "", err
	}
	reply, err := verifyCode.NewVerifyCodeClient(conn).GetVerifyCode(ctx, &verifyCode.GetVerifyCodeRequest{
		Length: 6,
		Type:   1,
	})
	if err != nil {
		return "", err
	}
	if reply.GetCode() == "" {
		return "", fmt.Errorf("验证码服务返回空码")
	}
	if err := cd.SetVerifyCode(ctx, telephone, reply.Code, biz.VerifyCodeLife); err != nil {
		return "", err
	}
	return reply.Code, nil
}

// SetVerifyCode 把验证码写进 Redis 1 号库。
// telephone 只作为键的一部分，函数不校验它是不是手机号。ex 的单位是秒，到点键自动删除。
func (cd *CustomerData) SetVerifyCode(ctx context.Context, telephone, code string, ex int64) error {
	return cd.data.Rdb.Set(ctx, "CVC:"+telephone, code, time.Duration(ex)*time.Second).Err()
}

// GetVerifyCode 读取 CVC:<手机号>。
// 键不存在、或 Redis 出错时都返回空字符串。登录把空串当成验证码不匹配，不区分这两种情况。
func (cd *CustomerData) GetVerifyCode(ctx context.Context, telephone string) string {
	code, err := cd.data.Rdb.Get(ctx, "CVC:"+telephone).Result()
	if err != nil {
		return ""
	}
	return code
}

// GetCustomerByTelephone 按手机号查乘客。
// 查到就返回这一行。确认是「没有这条记录」时插入一行，姓名、邮箱、微信保持 NULL，再把带主键的结构体返回。
// 其它数据库错误原样返回，不把它当成新用户。
func (cd *CustomerData) GetCustomerByTelephone(telephone string) (*biz.Customer, error) {
	customer := &biz.Customer{}
	result := cd.data.Mdb.Where("telephone = ?", telephone).First(customer)
	if result.Error == nil && customer.ID > 0 {
		return customer, nil
	}
	if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
		customer.Telephone = telephone
		if err := cd.data.Mdb.Create(customer).Error; err != nil {
			return nil, err
		}
		return customer, nil
	}
	return nil, result.Error
}

// GenerateTokenAndSave 签发乘客 JWT，并把签名结果写回这名乘客。
// c.ID 必须已经存在，它会变成令牌的 jti。duration 是有效期，secret 是 HS256 密钥。
// 标准声明里 Issuer 是 LaoMaDJ，Subject 是 customer-authentication，Audience 只有 customer。
// Save 会更新整行。表里只留最新一张令牌，所以上一次登录发出的令牌立刻不能再用。
func (cd *CustomerData) GenerateTokenAndSave(c *biz.Customer, duration time.Duration, secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "LaoMaDJ",
		Subject:   "customer-authentication",
		Audience:  []string{"customer"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        fmt.Sprintf("%d", c.ID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	c.Token = signedToken
	c.TokenCreatedAt = sql.NullTime{Time: time.Now(), Valid: true}
	if err := cd.data.Mdb.Save(c).Error; err != nil {
		return "", err
	}
	return signedToken, nil
}

// GetToken 按乘客主键读取当前令牌，给中间件和请求头里的 Bearer 比对。
// id 可以是字符串形式的主键，GORM 会把它当成 customers.id。查不到时返回错误。
func (cd *CustomerData) GetToken(id interface{}) (string, error) {
	c := &biz.Customer{}
	if err := cd.data.Mdb.First(c, id).Error; err != nil {
		return "", err
	}
	return c.Token, nil
}

// DelToken 按主键找到乘客，把 token 列写成空串，并把签发时间写成 NULL。
// 不删除乘客行。清空后，旧 JWT 即使还没过期也不能通过中间件。
func (cd *CustomerData) DelToken(id interface{}) error {
	c := &biz.Customer{}
	if err := cd.data.Mdb.First(c, id).Error; err != nil {
		return err
	}
	c.Token = ""
	c.TokenCreatedAt = sql.NullTime{Valid: false}
	return cd.data.Mdb.Save(c).Error
}
