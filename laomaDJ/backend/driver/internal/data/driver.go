// 本文件按手机号读写司机、验证码和令牌。库连接在 data.go。

package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"driver/api/verifyCode"
	"driver/internal/biz"
	"driver/internal/discover"

	"github.com/go-redis/redis/v9"
	"gorm.io/gorm"
)

// DriverData 是 biz.DriverInterface 的 MySQL / Redis 实现。
type DriverData struct {
	data *Data
}

// NewDriverInterface 供 wire 组装。返回接口，biz 看不到具体的表操作。
func NewDriverInterface(data *Data) biz.DriverInterface {
	return &DriverData{data: data}
}

// GetVerifyCode 向验证码服务要 6 位数字，写入 Redis 2 号库。
// 键是 DVC:<手机号>，过期时间与 biz.VerifyCodeLife 相同。Type 为 1，表示纯数字。
// Consul 可用时按名字 VerifyCode 发现，否则直连 127.0.0.1:9000。
func (dt *DriverData) GetVerifyCode(ctx context.Context, tel string) (string, error) {
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
	if err := dt.data.Rdb.Set(ctx, "DVC:"+tel, reply.GetCode(), time.Duration(biz.VerifyCodeLife)*time.Second).Err(); err != nil {
		return "", err
	}
	return reply.GetCode(), nil
}

// GetSavedVerifyCode 读取 DVC:<手机号>。
// 键不存在时返回「验证码不存在或已过期」，不把 redis.Nil 原样抛给页面。
func (dt *DriverData) GetSavedVerifyCode(ctx context.Context, tel string) (string, error) {
	code, err := dt.data.Rdb.Get(ctx, "DVC:"+tel).Result()
	if err == redis.Nil {
		return "", errors.New("验证码不存在或已过期")
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

// SaveToken 把 JWT 写到这台手机号的司机行。
// 查不到行时先调用 InitDriverInfo 建一条 stop 记录，再保存令牌。这样只登录、不单独提交手机号也能成功。
func (dt *DriverData) SaveToken(ctx context.Context, tel, token string) error {
	driver := biz.Driver{}
	err := dt.data.Mdb.Where("telephone = ?", tel).First(&driver).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		created, cerr := dt.InitDriverInfo(ctx, tel)
		if cerr != nil {
			return cerr
		}
		driver = *created
	} else if err != nil {
		return err
	}
	driver.Token = sql.NullString{String: token, Valid: true}
	return dt.data.Mdb.Save(&driver).Error
}

// ClearToken 把该手机号的 token 列更新为 NULL。找不到行时 GORM 的更新影响 0 行，不返回错误。
func (dt *DriverData) ClearToken(ctx context.Context, tel string) error {
	return dt.data.Mdb.Model(&biz.Driver{}).Where("telephone = ?", tel).Updates(map[string]any{
		"token": nil,
	}).Error
}

// GetToken 按手机号读取当前令牌字符串。没有这个司机时返回查询错误。令牌列是 NULL 时返回空串。
func (dt *DriverData) GetToken(ctx context.Context, tel string) (string, error) {
	driver := biz.Driver{}
	if err := dt.data.Mdb.Where("telephone = ?", tel).First(&driver).Error; err != nil {
		return "", err
	}
	return driver.Token.String, nil
}

// InitDriverInfo 按手机号查找。找到就返回原行，不改状态。
// 确认没有记录时插入一行，状态写成 stop，表示已登记但还不能听单。
func (dt *DriverData) InitDriverInfo(ctx context.Context, tel string) (*biz.Driver, error) {
	driver := biz.Driver{}
	err := dt.data.Mdb.Where("telephone = ?", tel).First(&driver).Error
	if err == nil {
		return &driver, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	driver.Telephone = tel
	driver.Status = sql.NullString{String: biz.DriverStatusStop, Valid: true}
	if err := dt.data.Mdb.Create(&driver).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

// FetchInfoByTel 按手机号读出整行司机。中间件把它放进 context，后续接口可以取出，当前业务还没有使用这份上下文值。
func (dt *DriverData) FetchInfoByTel(ctx context.Context, tel string) (*biz.Driver, error) {
	driver := &biz.Driver{}
	if err := dt.data.Mdb.Where("telephone = ?", tel).First(driver).Error; err != nil {
		return nil, err
	}
	return driver, nil
}
