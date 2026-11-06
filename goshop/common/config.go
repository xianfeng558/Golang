// Package common 是五个微服务共用的工具箱：
// 读配置、连 MySQL、连 Redis、发统一 JSON、做链路追踪和监控。
// 各服务通过 import 路径 git.mashibing.com/msb_47094/shopping-comm 使用它。
// go.work 把这个目录映射成那个模块，所以改这里立刻对全部服务生效。
package common

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote" // 让 viper 认识 consul 这种远程配置源
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// ConsulIp 是注册中心所在机器。本机学习时用 127.0.0.1。
	// 课程原代码写的是局域网 IP，换电脑就会连不上。
	ConsulIp = "127.0.0.1"
	// ConsulStr 是 viper 读远程配置时用的 HTTP 地址。
	ConsulStr = "http://" + ConsulIp + ":8500"
	// ConsulReistStr 是 go-micro 注册服务时用的 host:port。名字里的 Reist 是课程原文的拼写，保留以免改调用处。
	ConsulReistStr = ConsulIp + ":8500"
	// DtmServer 是分布式事务管理器 DTM 的地址。没安装 DTM 时，下单接口会报提交失败，其它接口不受影响。
	DtmServer = "http://" + ConsulIp + ":36789/api/dtmsvr"
	// QSIp 本机回环地址。熔断器的监控页绑在这上面。
	QSIp = "127.0.0.1"
	// CartHTTP 购物车 HTTP 网关。DTM 要回调“减库存 / 加购物车”时打到这个地址。
	CartHTTP = "http://" + QSIp + ":6668"
	// QSBusi 订单 HTTP 网关。DTM 回调“删购物车 / 写订单”时打到这个地址。
	QSBusi = "http://" + QSIp + ":6669"
	// 下面四个字符串是 Consul 配置中心里的键名。Consul 没开时用各服务 config 目录里的同名 JSON。
	ProductFileKey = "mysql-product"
	TradeFileKey   = "mysql-trade"
	UserFileKey    = "mysql-user"
	RedisFileKey   = "redis"
	// QPS 是 go-micro 限流器每秒放行的请求数。超过就直接拒绝，用来保护数据库。
	QPS = 100
)

// ConsulReachable 用 1.5 秒判断本机 8500 端口有没有 Consul。
// 没有就让服务改用内嵌 JSON 和 mDNS，这样不装 Consul 也能把项目跑起来。
func ConsulReachable() bool {
	conn, err := net.DialTimeout("tcp", ConsulReistStr, 1500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// ParseJSON 把一段 JSON 字节读成 viper。服务启动时用它解析 //go:embed 嵌进来的配置文件。
func ParseJSON(body []byte) (*viper.Viper, error) {
	conf := viper.New()
	conf.SetConfigType("json")
	if err := conf.ReadConfig(bytes.NewReader(body)); err != nil {
		return nil, err
	}
	return conf, nil
}

// GetConsulConfig 从 Consul 的 KV 里读一份 JSON。
// url 例子：http://127.0.0.1:8500 ，fileKey 例子：mysql-user。
// 读失败时仍然返回一个空的 viper，调用方要自己检查 host / addr 是否为空。
func GetConsulConfig(url string, fileKey string) (*viper.Viper, error) {
	conf := viper.New()
	conf.AddRemoteProvider("consul", url, fileKey)
	conf.SetConfigType("json")
	if err := conf.ReadRemoteConfig(); err != nil {
		log.Println("viper conf err :", err)
		return conf, err
	}
	return conf, nil
}

// EnsureDatabase 先连到 MySQL 服务器（不指定库名），再执行 CREATE DATABASE IF NOT EXISTS。
// 这样第一次启动不用手动建库。
func EnsureDatabase(conf *viper.Viper) error {
	if conf == nil {
		return fmt.Errorf("mysql config is nil")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		conf.GetString("user"), conf.GetString("pwd"), conf.GetString("host"), conf.GetString("port"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS `" + conf.GetString("database") + "` DEFAULT CHARACTER SET utf8mb4")
	return err
}

// GetMysqlFromConsul 用 viper 里的 host/port/user/pwd/database 打开 GORM。
// 函数名带 Consul，是因为课程最初只从 Consul 拿配置；现在同一份结构的本地 JSON 也能传进来。
func GetMysqlFromConsul(vip *viper.Viper) (*gorm.DB, error) {
	if vip == nil || vip.GetString("host") == "" {
		return nil, fmt.Errorf("mysql config is empty")
	}
	// 慢 SQL（超过 1 秒）会打到控制台，学习时能看见 GORM 实际执行的语句。
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		vip.GetString("user"), vip.GetString("pwd"), vip.GetString("host"), vip.GetString("port"), vip.GetString("database"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}
	// 本机 MySQL 偶尔会掐掉空闲连接。限制连接寿命，下次取连接时会新建，而不是一直拿着一条死连接。
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	if err = sqlDB.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// GetRedisFromConsul 按配置创建 Redis 客户端，并 Ping 一次。
// 连不上就返回错误，避免服务“看起来启动了，登录时才发现缓存不可用”。
func GetRedisFromConsul(vip *viper.Viper) (*redis.Client, error) {
	if vip == nil || vip.GetString("addr") == "" {
		return nil, fmt.Errorf("redis config is empty")
	}
	red := redis.NewClient(&redis.Options{
		Addr:         vip.GetString("addr"),
		Password:     vip.GetString("password"),
		DB:           vip.GetInt("DB"),
		PoolSize:     vip.GetInt("poolSize"),
		MinIdleConns: vip.GetInt("minIdleConn"),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := red.Ping(ctx).Err(); err != nil {
		_ = red.Close()
		return nil, fmt.Errorf("redis ping %s: %w", vip.GetString("addr"), err)
	}
	return red, nil
}

// SetUserToken 把登录令牌写进 Redis。timeTTL 到期后键自动消失，相当于登录过期。
func SetUserToken(red *redis.Client, key string, val []byte, timeTTL time.Duration) {
	if red == nil {
		log.Print("SetUserToken skipped, redis is nil")
		return
	}
	red.Set(context.Background(), key, val, timeTTL)
}

// SetOrderToken 和 SetUserToken 一样，只是值用字符串。当前下单流程没有调用它。
func SetOrderToken(red *redis.Client, key string, val string, timeTTL time.Duration) {
	if red == nil {
		return
	}
	red.Set(context.Background(), key, val, timeTTL)
}

// GetUserToken 按键取出令牌。键不存在时 Redis 返回 redis.Nil，这里记一条日志并返回空字符串。
func GetUserToken(red *redis.Client, key string) string {
	if red == nil {
		return ""
	}
	res, err := red.Get(context.Background(), key).Result()
	if err != nil {
		log.Print("GetUserToken  err  ", err)
		return ""
	}
	return res
}
