// Package model 是用户表。GORM 把这个结构体映射成 user_center.user。
package model

import "time"

// User 的 ID 不写 tag 时，GORM 默认把它当成主键，列名是 id。
// 验证码登录不会校验 Password，这个字段只是表结构的一部分。
// User 的每个字段都会变成 user 表的一列。
// autoCreateTime 表示插入时如果没填时间，GORM 用当前时间。
type User struct {
	ID            int32     // 主键。GORM 看到名字是 ID 就知道它是主键
	Avatar        string    `gorm:"default:'https://msb-edu-dev.oss-cn-beijing.aliyuncs.com/default-headimg.png'"`
	ClientId      int32     `gorm:"default:1"` // 来自哪个客户端，课程登录时要求传 0
	Nickname      string    `gorm:"default:'随机名称'"`
	Phone         string    // 登录用的手机号。没有 unique 标签，课程没有禁止重复
	Password      string    `gorm:"default:'1234'"` // 当前登录流程不校验它
	SystemId      string    `gorm:"default:1"`
	LastLoginTime time.Time `gorm:"autoCreateTime"`
	CreateTime    time.Time `gorm:"autoCreateTime"`
	IsDeleted     int32     `gorm:"default:0"` // 0 未删除。这里用整数，商品表用的是 bool，两套写法都来自课程
	UnionId       string    `gorm:"default:'1'"`
}

// TableName 固定表名是 user。不写的话 GORM 会使用 users。
func (table *User) TableName() string {
	return "user"
}
