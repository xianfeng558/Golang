// Package models 是表和连接。
// 用户、好友、群落在 MySQL；在线标记和聊天记录落在 Redis。
// WebSocket 连接只活在这个进程的内存里，所以同一条消息要在本机找到收件人，再写进 Redis。
package models

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"ginchat/utils"
	"log"
	"time"
)

// UserBasic 对应表 user_basic，一行就是一个注册用户。
// 内嵌的 Base 提供 ID、CreatedAt、UpdatedAt、DeletedAt。JSON 里的主键名是 ID，登录页读的是 data.ID。
// PassWord 和 Salt 带 json:"-"，注册和登录的响应里不会出现摘要。
// Identity 是登录后写上的令牌，WebSocket 的 token 必须和它一致。
type UserBasic struct {
	Base
	Name          string    `gorm:"comment:登录名"`
	PassWord      string    `gorm:"comment:密码摘要，MD5(明文加盐)" json:"-"`
	Phone         string    `gorm:"comment:手机号" valid:"optional,matches(^1[3-9]\\d{9}$)"`
	Email         string    `gorm:"comment:邮箱" valid:"optional,email"`
	Avatar        string    `gorm:"comment:头像地址"`
	Identity      string    `gorm:"comment:登录令牌，WebSocket 的 token 必须等于它"`
	ClientIp      string    `gorm:"comment:预留的客户端 IP"`
	ClientPort    string    `gorm:"comment:预留的客户端端口"`
	Salt          string    `gorm:"comment:6 位数字盐" json:"-"`
	LoginTime     time.Time `gorm:"comment:最近一次签发令牌的时间"`
	HeartbeatTime time.Time `gorm:"comment:注册时记下的时间，连接心跳不回写这一列"`
	LoginOutTime  time.Time `gorm:"column:login_out_time;comment:预留的退出时间" json:"login_out_time"`
	IsLogout      bool      `gorm:"comment:预留的登出标记"`
	DeviceInfo    string    `gorm:"comment:预留的设备信息"`
}

func (UserBasic) TableComment() string { return "注册用户" }

// TableName 固定表名。不写的话 GORM 可能把它变成 user_basics。
func (table *UserBasic) TableName() string {
	return "user_basic"
}

// GetUserList 返回全部未软删除的用户。
// 切片用 var 声明，长度是 0。不要预分配 10 个空位，否则查不到人时也会带出一串 nil。
// 聊天页不用这个接口，它留给管理列表。
func GetUserList() ([]*UserBasic, error) {
	var data []*UserBasic
	err := utils.DB.Find(&data).Error
	return data, err
}

// FindUserByName 按用户名找一个人。
// 用 Find 而不是 First：没有这行时 First 会打出 record not found，注册新用户看起来像出错。
// 没找到时返回的 ID 是 0，调用方靠 ID 判断。
func FindUserByName(name string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("name = ?", name).Limit(1).Find(&user)
	return user
}

// FindUserByPhone 按手机号查找。
// 登录页现在只认用户名。这个函数留给以后改成手机号登录，没找到时 ID 为 0。
func FindUserByPhone(phone string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("phone = ?", phone).Limit(1).Find(&user)
	return user
}

// FindUserByEmail 按邮箱查找。当前页面不用，没找到时 ID 为 0。
func FindUserByEmail(email string) UserBasic {
	user := UserBasic{}
	utils.DB.Where("email = ?", email).Limit(1).Find(&user)
	return user
}

// FindByID 按主键查找。
// id 为 0 时直接返回空用户，避免扫出一条意外的数据。没找到时 ID 仍是 0。
func FindByID(id uint) UserBasic {
	user := UserBasic{}
	if id == 0 {
		return user
	}
	utils.DB.Where("id = ?", id).Limit(1).Find(&user)
	return user
}

// CreateUser 插入一行。
// user 必须是指针：GORM 把自增 ID 写回这个对象，注册接口才能在响应里带上 ID。
func CreateUser(user *UserBasic) error {
	return utils.DB.Create(user).Error
}

// DeleteUser 按主键软删除。
// gorm.Model 带了 DeletedAt，所以行还在库里，只是 deleted_at 被填上。以后的 Find 会自动滤掉它。
// ID 为 0 时拒绝执行，防止 GORM 误删整表。
func DeleteUser(user UserBasic) error {
	if user.ID == 0 {
		return fmt.Errorf("缺少用户 ID")
	}
	return utils.DB.Delete(&user).Error
}

// UpdateProfile 只改传入的非空项。
// 空字符串不会写回数据库，所以改资料时不填手机、邮箱、密码，原来的值还在。
// 密码一旦有值，就用这个用户原来的盐重新做摘要。老用户没有盐时补一份新的。
func UpdateProfile(user UserBasic) error {
	if user.ID == 0 {
		return fmt.Errorf("缺少用户 ID")
	}
	old := FindByID(user.ID)
	if old.ID == 0 {
		return fmt.Errorf("用户不存在")
	}
	updates := map[string]interface{}{}
	if user.Name != "" {
		updates["name"] = user.Name
	}
	if user.Phone != "" {
		updates["phone"] = user.Phone
	}
	if user.Email != "" {
		updates["email"] = user.Email
	}
	if user.Avatar != "" {
		updates["avatar"] = user.Avatar
	}
	if user.PassWord != "" {
		salt := old.Salt
		if salt == "" {
			salt = NewSalt()
			updates["salt"] = salt
		}
		updates["pass_word"] = utils.MakePassword(user.PassWord, salt)
	}
	if len(updates) == 0 {
		return nil
	}
	return utils.DB.Model(&old).Updates(updates).Error
}

// IssueToken 签发登录令牌并写进 identity。
// 材料是「用户 ID-当前纳秒」的大写 MD5。带上用户 ID，两个用户同一纳秒登录也不会得到同一个令牌。
// 令牌会同时放进 user.Identity，调用方把这个结构体编成 JSON 后，页面才能读到 data.Identity。
// 它不是密码，也没有单独的过期列；连接时只比较字符串是否还等于库里的值。
func IssueToken(user *UserBasic) error {
	if user == nil || user.ID == 0 {
		return fmt.Errorf("无法签发令牌")
	}
	token := utils.MD5Encode(fmt.Sprintf("%d-%d", user.ID, time.Now().UnixNano()))
	if err := utils.DB.Model(user).Update("identity", token).Error; err != nil {
		return err
	}
	user.Identity = token
	user.LoginTime = time.Now()
	return nil
}

// NewSalt 生成 6 位数字盐，不足 6 位时左边补 0。
// 优先用 crypto/rand。读随机数失败时才退回当前纳秒，避免注册因为盐生成失败而中断。
// 注册时每人一份，以后改密码还用这一份，否则旧摘要会对不上。
func NewSalt() string {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", binary.BigEndian.Uint32(buf[:])%1000000)
}

// SetUserOnlineInfo 把在线标记写进 Redis。
// key 一般是 online_用户ID，val 是这条连接的远端地址，ttl 到期后投递逻辑会把人看成离线。
// 心跳到来时用同一个键再写一次，等于续期。Redis 没初始化时只打日志，不让调用方崩溃。
func SetUserOnlineInfo(key string, val []byte, ttl time.Duration) {
	if utils.Red == nil {
		log.Println("SetUserOnlineInfo skipped, redis is nil")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := utils.Red.Set(ctx, key, val, ttl).Err(); err != nil {
		log.Println("set online", key, err)
	}
}
