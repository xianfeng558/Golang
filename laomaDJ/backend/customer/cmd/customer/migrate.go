package main

import (
	"database/sql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

// 定义模型
type Customer struct {
	gorm.Model
	Telephone      string       `gorm:"type:varchar(16);uniqueIndex" json:"telephone"`
	Token          string       `gorm:"type:varchar(4096);" json:"token"`
	TokenCreatedAt sql.NullTime `gorm:"" json:"token_created_at"`
	Name           string       `gorm:"type:varchar(255);uniqueIndex" json:"name"`
	Email          string       `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Wechat         string       `gorm:"type:varchar(255);uniqueIndex" json:"wechat"`
}

// 初始时执行
func init() {
	// 一，建立数据库连接
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:mashibing@tcp(127.0.0.1:3306)/laomadj_customer?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	// 表创建，结构更新
	if err := db.AutoMigrate(&Customer{}); err != nil {
		log.Fatalln(err)
	}
}
