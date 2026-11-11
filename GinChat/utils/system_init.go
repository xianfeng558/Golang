// Package utils 负责配置、MySQL、Redis，以及两套 JSON 响应。
// 业务包通过 utils.DB 和 utils.Red 使用连接，自己不再写地址。
package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// DB 是进程里唯一的 GORM 连接。InitMySQL 成功之后才可以查询。
	DB *gorm.DB
	// Red 是进程里唯一的 Redis 客户端。在线状态和聊天记录都走它。
	Red *redis.Client
)

// PublishKey 是演示用的 Redis 发布订阅频道。聊天页不走这里，它走 WebSocket 直连。
const PublishKey = "websocket"

// InitConfig 读取 config/app.yml。工作目录必须是 GinChat 根目录，因为路径是相对的。
func InitConfig() error {
	viper.SetConfigName("app")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取 config/app.yml: %w", err)
	}
	log.Println("config loaded", viper.ConfigFileUsed())
	return nil
}

// mysqlDSN 拼出 GORM 用的连接串。
// database 传空字符串时连到服务器本身，用来执行 CREATE DATABASE；传库名时才进入 ginchat。
// parseTime 让 DATETIME 能扫进 time.Time，loc=Local 用本机时区，所以读出来的时间带 +08:00。
func mysqlDSN(database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.pwd"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		database,
	)
}

// InitMySQL 先确保库存在，再打开 GORM。账号或地址不对时直接返回错误，避免后面第一次查询才炸。
func InitMySQL() error {
	dbName := viper.GetString("mysql.database")
	if dbName == "" {
		return fmt.Errorf("mysql.database 为空")
	}
	server, err := gorm.Open(mysql.Open(mysqlDSN("")), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接 MySQL: %w", err)
	}
	if err = server.Exec("CREATE DATABASE IF NOT EXISTS `" + dbName + "` DEFAULT CHARACTER SET utf8mb4").Error; err != nil {
		return fmt.Errorf("创建数据库 %s: %w", dbName, err)
	}
	sqlServer, err := server.DB()
	if err != nil {
		return err
	}
	_ = sqlServer.Close()

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	DB, err = gorm.Open(mysql.Open(mysqlDSN(dbName)), &gorm.Config{Logger: newLogger})
	if err != nil {
		return fmt.Errorf("打开数据库 %s: %w", dbName, err)
	}
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	// 本机 MySQL 会掐掉闲太久的连接。寿命限制在 30 秒，下次取连接时新建，避免拿着一条已经死掉的连接。
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Second)
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping MySQL: %w", err)
	}
	log.Println("MySQL ready", dbName)
	return nil
}

// AutoMigrate 按模型建表、补列。已经有的列和数据留着。
// 模型如果实现了 TableComment，新建表时会把这句话写成表注释。
func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("MySQL 还没有初始化")
	}
	for _, model := range models {
		session := DB
		if commented, ok := model.(interface{ TableComment() string }); ok {
			if comment := commented.TableComment(); comment != "" {
				session = DB.Set("gorm:table_options", "COMMENT='"+comment+"'")
			}
		}
		if err := session.AutoMigrate(model); err != nil {
			return err
		}
	}
	return nil
}

// InitRedis 按配置创建客户端并 Ping。聊天记录依赖 Redis，连不上就不要假装服务可用。
func InitRedis() error {
	addr := viper.GetString("redis.addr")
	if addr == "" {
		return fmt.Errorf("redis.addr 为空")
	}
	Red = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     viper.GetString("redis.password"),
		DB:           viper.GetInt("redis.DB"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConn"),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := Red.Ping(ctx).Err(); err != nil {
		_ = Red.Close()
		Red = nil
		return fmt.Errorf("redis ping %s: %w", addr, err)
	}
	log.Println("Redis ready", addr)
	return nil
}

// Publish 往 Redis 频道发一条字符串。演示接口 /user/sendMsg 用它看发布订阅。
func Publish(ctx context.Context, channel string, msg string) error {
	if Red == nil {
		return fmt.Errorf("redis 未初始化")
	}
	return Red.Publish(ctx, channel, msg).Err()
}

// Subscribe 订一次频道并阻塞到下一条消息。调用方每调一次都会新建一个订阅。
func Subscribe(ctx context.Context, channel string) (string, error) {
	if Red == nil {
		return "", fmt.Errorf("redis 未初始化")
	}
	sub := Red.Subscribe(ctx, channel)
	defer sub.Close()
	msg, err := sub.ReceiveMessage(ctx)
	if err != nil {
		return "", err
	}
	return msg.Payload, nil
}
