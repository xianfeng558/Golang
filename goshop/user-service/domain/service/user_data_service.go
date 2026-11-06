// Package service 是业务层。它不写 SQL，只把 repository 的能力再暴露给 handler。
// handler 依赖接口，以后可以换成别的存储实现。
package service

import (
	"time"

	"gouser/domain/model"
	"gouser/domain/repository"
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

func (u *UserDataService) Login(clientId int32, phone string, systemId int32, verificationCode string) (*model.User, error) {
	return u.userRepository.Login(clientId, phone, systemId, verificationCode)
}

func (u *UserDataService) SetUserToken(key string, val []byte, timeTTL time.Duration) {
	u.userRepository.SetUserToken(key, val, timeTTL)
}

func (u *UserDataService) GetUserToken(key string) string {
	return u.userRepository.GetUserToken(key)
}
