// Package service 是业务层。它不关心 SQL 怎么写，只把 repository 的能力再暴露出去。
// 这样 handler 依赖接口，测试时可以换成假的 repository。
package service

import (
	"gouser/domain/model"
	"gouser/domain/repository"
	"time"
)

type IUserDataService interface {
	Login(int32, string, int32, string) (*model.User, error)
	SetUserToken(key string, val []byte, timeTTL time.Duration)
	GetUserToken(key string) string
}
type UserDataService struct {
	userRepository repository.IUserRepository
}

func NewUserDataService(userRepository repository.IUserRepository) IUserDataService {
	return &UserDataService{userRepository: userRepository}
}

// 重写接口方法
func (u *UserDataService) Login(clientId int32, phone string, systemId int32, verificationCode string) (user *model.User, err error) {

	return u.userRepository.Login(clientId, phone, systemId, verificationCode)
}

/*
	clientId, _ := strconv.Atoi(c.Request.FormValue("clientId"))

phone := c.Request.FormValue("phone")
systemId, _ := strconv.Atoi(c.Request.FormValue("systemId"))
verificationCode := c.Request.FormValue("verificationCode")
*/
func (u *UserDataService) SetUserToken(key string, val []byte, timeTTL time.Duration) {
	u.userRepository.SetUserToken(key, val, timeTTL)
}

func (u *UserDataService) GetUserToken(key string) string {
	return u.userRepository.GetUserToken(key)
}
