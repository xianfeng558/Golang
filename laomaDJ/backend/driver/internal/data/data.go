// Package data 连接司机库和 Redis 2 号库。
// 验证码键是 DVC:<手机号>。令牌写在 drivers.token。没有这行司机时，登录会先建成 stop。
package data

import (
	"context"
	"database/sql"
	"driver/internal/biz"
	"driver/internal/conf"
	"fmt"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v9"
	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet 是 data 层交给 wire 的构造函数。
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewDriverInterface)

// Data 持有司机服务的 MySQL、Redis，以及模板留下的服务配置。
type Data struct {
	Mdb *gorm.DB
	Rdb *redis.Client
	cs  *conf.Service
}

// NewData 连接 Redis（2 号库）和 MySQL，库不存在时创建 laomadj_driver。
func NewData(c *conf.Data, cs *conf.Service, logger log.Logger) (*Data, func(), error) {
	if c.GetRedis() == nil || c.Redis.GetAddr() == "" {
		return nil, nil, fmt.Errorf("redis config is empty")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		DB:           2,
		DialTimeout:  time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	})
	pingCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		_ = rdb.Close()
		return nil, nil, fmt.Errorf("redis: %w", err)
	}

	db, err := openMySQL(c.GetDatabase().GetSource())
	if err != nil {
		_ = rdb.Close()
		return nil, nil, err
	}
	if err := db.AutoMigrate(&biz.Driver{}); err != nil {
		_ = rdb.Close()
		return nil, nil, fmt.Errorf("migrate driver: %w", err)
	}
	sqlDB, _ := db.DB()
	cleanup := func() {
		_ = rdb.Close()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{Mdb: db, Rdb: rdb, cs: cs}, cleanup, nil
}

// openMySQL 先不带库名连上服务器，执行 CREATE DATABASE IF NOT EXISTS，再打开 laomadj_driver。
// 库名来自 DSN。空名或带引号会拒绝，避免把配置拼进 SQL。
// 最多 10 条连接，空闲 2 条，每条最多用 30 秒。本机 MySQL 会掐掉更老的空闲连接，下次查询会新建。
func openMySQL(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("mysql dsn is empty")
	}
	cfg, err := mysqldriver.ParseDSN(dsn)
	if err != nil {
		return nil, err
	}
	name := cfg.DBName
	if name == "" || strings.ContainsAny(name, "`;'\"") {
		return nil, fmt.Errorf("invalid mysql database name")
	}
	cfg.DBName = ""
	admin, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}
	defer admin.Close()
	if err := admin.Ping(); err != nil {
		return nil, err
	}
	if _, err := admin.Exec("CREATE DATABASE IF NOT EXISTS `" + name + "` DEFAULT CHARACTER SET utf8mb4"); err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	return db, nil
}
