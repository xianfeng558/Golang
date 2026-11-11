// Package model 是用户表。GORM 把这个结构体映射成 user_center.user。
package model

import "time"

// User 的 ID 不写 tag 时，GORM 默认把它当成主键，列名是 id。
// 验证码登录不会校验 Password，这个字段只是表结构的一部分。
// User 的每个字段都会变成 user 表的一列。
// autoCreateTime 表示插入时如果没填时间，GORM 用当前时间。
type User struct {
	ID            int32     `gorm:"primaryKey;comment:用户主键"`
	Avatar        string    `gorm:"default:'https://msb-edu-dev.oss-cn-beijing.aliyuncs.com/default-headimg.png';comment:头像地址"`
	ClientId      int32     `gorm:"default:1;comment:客户端编号，登录接口要求传 0"`
	Nickname      string    `gorm:"default:'随机名称';comment:昵称"`
	Phone         string    `gorm:"comment:登录手机号"`
	Password      string    `gorm:"default:'1234';comment:当前登录不校验密码"`
	SystemId      string    `gorm:"default:1;comment:系统编号，登录接口要求传 0"`
	LastLoginTime time.Time `gorm:"autoCreateTime;comment:最近登录时间"`
	CreateTime    time.Time `gorm:"autoCreateTime;comment:注册时间"`
	IsDeleted     int32     `gorm:"default:0;comment:0 未删除"`
	UnionId       string    `gorm:"default:'1';comment:预留的联合编号"`
}

func (User) TableComment() string { return "用户。令牌在 Redis，键是用户 ID" }

// TableName 固定表名是 user。不写的话 GORM 会使用 users。
func (table *User) TableName() string {
	return "user"
}
