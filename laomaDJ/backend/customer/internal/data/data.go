// 本文件在进程启动时打开 Redis 1 号库和乘客库，并按 biz.Customer 建表。

package data

import (
	"context"
	"customer/internal/biz"
	"customer/internal/conf"
	"database/sql"
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
// NewGreeterRepo 属于 Kratos 模板，代驾业务不调用它，删掉会导致现有的 wire_gen.go 对不上。
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewCustomerData)

// Data 持有乘客服务要用的两个客户端。
// Rdb 指向 Redis 1 号库，和司机使用的 2 号库分开，避免验证码键混在一起。
// Mdb 指向库 laomadj_customer。
type Data struct {
	Rdb *redis.Client
	Mdb *gorm.DB
}

// NewData 按配置连接 Redis 和 MySQL。
// Redis 连的是 c.Redis.Addr 的 1 号库，启动时 Ping 一次，连不上就返回错误，避免服务看起来起来了、登录时才失败。
// MySQL 会先建库 laomadj_customer，再按 biz.Customer 建表或补列。
// 返回的 cleanup 在进程退出时关闭两个客户端。
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	if c.GetRedis() == nil || c.Redis.GetAddr() == "" {
		return nil, nil, fmt.Errorf("redis config is empty")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		DB:           1,
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
	if err := db.AutoMigrate(&biz.Customer{}, &biz.TripOrder{}); err != nil {
		_ = rdb.Close()
		return nil, nil, fmt.Errorf("migrate customer: %w", err)
	}

	sqlDB, _ := db.DB()
	cleanup := func() {
		_ = rdb.Close()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{Rdb: rdb, Mdb: db}, cleanup, nil
}

// openMySQL 先不带库名连上服务器，执行 CREATE DATABASE IF NOT EXISTS，再打开目标库。
// 库名来自 DSN，拒绝空名和带引号的名字，避免把配置拼进 SQL 时破坏语句。
// 连接寿命限制为 30 秒。本机 MySQL 会掐掉空闲连接，下次取连接时会新建，而不是继续用一条已经断开的连接。
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
	// 本机 MySQL 会掐掉空闲连接。限制寿命后，下次查询会新建连接。
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	return db, nil
}
