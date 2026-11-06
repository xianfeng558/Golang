// Package repository 直接操作 MySQL 和 Redis。
// 登录规则是课程里写死的：clientId、systemId 都是 0，验证码必须是 6666。
// 满足条件就按手机号查找，找不到就插入一行，相当于“验证码通过即注册”。
package repository

import (
	"errors"
	"fmt"
	common "git.mashibing.com/msb_47094/shopping-comm"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"gouser/domain/model"
	"time"
)

/**
 int32 clientId = 1;
    string phone = 2;
    int32 systemId = 3;
    string verificationCode = 4;
**/
//接口
type IUserRepository interface {
	Login(int32, string, int32, string) (*model.User, error)
	SetUserToken(key string, val []byte, timeTTL time.Duration)
	GetUserToken(key string) string
}

// 创建实例
func NewUserRepository(db *gorm.DB, red *redis.Client) IUserRepository {
	return &UserRepository{mysqlDB: db, red: red}
}

// 数据DB
type UserRepository struct {
	mysqlDB *gorm.DB
	red     *redis.Client
}

// 重写接口方法
func (u *UserRepository) Login(clientId int32, phone string, systemId int32, verificationCode string) (user *model.User, err error) {
	user = &model.User{}
	if clientId == 0 && systemId == 0 && verificationCode == "6666" {
		u.mysqlDB.Where("phone = ? ", phone).Find(user)
		//未找到就注册一个
		if user.ID == 0 {
			user.Phone = phone
			u.mysqlDB.Create(&user)
			//u.mysqlDB.Select("Nickname", "Avatar", "Phone", "ClientId").Create(&user)
		}
		return user, nil
		//return user, u.mysqlDB.Where("phone = ? ", phone).Find(user).Error
	} else {
		return user, errors.New("参数不匹配")
	}
}

// SetUserToken 写两把 Redis 键，方便入门时对上号：
// 1. 原始用户 ID，例如 "2"，请求头 Uuid 直接填 2
// 2. 课程原来的二进制形式，例如 2 会写成 "10"，旧调用方式仍然有效
func (u *UserRepository) SetUserToken(key string, val []byte, timeTTL time.Duration) {
	fmt.Println("user token redis key:", key)
	common.SetUserToken(u.red, key, val, timeTTL)
	binKey := common.ConverToBinary(common.ToInt(key))
	if binKey != key {
		common.SetUserToken(u.red, binKey, val, timeTTL)
	}
}

func (u *UserRepository) GetUserToken(key string) string {
	return common.GetUserToken(u.red, key)
}
