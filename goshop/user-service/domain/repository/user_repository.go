// Package repository 直接操作 MySQL 和 Redis。
// 登录规则写死为：clientId、systemId 都是 0，验证码必须是 6666。
// 条件满足就按手机号查找，找不到就插入一行，相当于验证码通过即注册。
package repository

import (
	"errors"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gouser/domain/model"
)

// IUserRepository 是用户服务需要的三件事：登录、写令牌、读令牌。
type IUserRepository interface {
	Login(int32, string, int32, string) (*model.User, error)
	SetUserToken(key string, val []byte, timeTTL time.Duration)
	GetUserToken(key string) string
}

// NewUserRepository 同时拿着 MySQL 和 Redis。登录查表，令牌进缓存。
func NewUserRepository(db *gorm.DB, red *redis.Client) IUserRepository {
	return &UserRepository{mysqlDB: db, red: red}
}

type UserRepository struct {
	mysqlDB *gorm.DB
	red     *redis.Client
}

// Login 校验三个固定条件。数据库查询或插入失败时把错误交回去，不再当成登录成功。
func (u *UserRepository) Login(clientId int32, phone string, systemId int32, verificationCode string) (*model.User, error) {
	user := &model.User{}
	if clientId != 0 || systemId != 0 || verificationCode != "6666" {
		return user, errors.New("参数不匹配")
	}
	if err := u.mysqlDB.Where("phone = ?", phone).Find(user).Error; err != nil {
		return nil, err
	}
	if user.ID == 0 {
		user.Phone = phone
		if err := u.mysqlDB.Create(user).Error; err != nil {
			return nil, err
		}
	}
	return user, nil
}

// SetUserToken 写两把 Redis 键，过期时间由调用方传入，登录时是 1 小时。
// 十进制用户 ID（例如 "2"）和二进制形式（2 写成 "10"）都写一份，请求头填 2 或 10 都能通过。
func (u *UserRepository) SetUserToken(key string, val []byte, timeTTL time.Duration) {
	common.SetUserToken(u.red, key, val, timeTTL)
	binKey := common.ConverToBinary(common.ToInt(key))
	if binKey != key {
		common.SetUserToken(u.red, binKey, val, timeTTL)
	}
}

func (u *UserRepository) GetUserToken(key string) string {
	return common.GetUserToken(u.red, key)
}
