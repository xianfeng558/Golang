// 命令：在 user-service 目录执行 go run .
// 作用：启动名为 shop-user 的 RPC 服务，监听 8081。
// 浏览器不直接访问 8081，而是访问 client 目录里 8088 端口的 HTTP 网关。
package main

import (
	"embed" // 编译时把 JSON 配置打进可执行文件
	"log"   // 打印启动日志。Fatal 会打印后让进程退出
	"time"  // 注册中心的心跳间隔

	// common 是五个服务共用的工具。引号里是模块路径，go.work 会指到本地的 goshop/common。
	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4" // 注册中心。8500 没开时不会用它
	// 下划线表示只执行这个包的初始化。GORM 连 MySQL 前，驱动要先注册自己。
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper" // 读 JSON。后面用 GetString("host")、GetString("addr") 取字段
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"gouser/domain/model"      // 用户表结构，AutoMigrate 靠它建表
	"gouser/domain/repository" // 查 MySQL、写 Redis 令牌
	"gouser/domain/service"
	"gouser/handler" // Login 和 GetUserToken 的实现
	"gouser/proto"   // user.proto 生成的注册函数
)

// go:embed 必须紧挨着下面的变量。两个 JSON 会一起打进程序。
// 改完配置要重新 go run，只改文件不重启不会生效。
//
//go:embed config/mysql-user.json config/redis.json
var localConfig embed.FS

// main 的顺序：读配置、组装 RPC、建库建表、连 Redis、注册登录和校验令牌，最后 Run 堵住不返回。
func main() {
	// 三个返回值：MySQL 配置、Redis 配置、可选的 Consul。没有 Consul 时 reg 是 nil。
	mysqlCfg, redisCfg, reg := loadConfig()

	// micro.Option 是一条启动参数。切片用来收集多条，有 Consul 时再 append。
	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),      // 注册信息 30 秒过期，进程挂了会被摘掉
		micro.RegisterInterval(time.Second * 30), // 每 30 秒续一次注册
		micro.Name("shop-user"),                  // 购物车和订单靠这个名字来调用，不写死 IP
		micro.Address(":8081"),                   // 冒号表示监听本机所有网卡的 8081
		micro.Version("v1"),
	}
	// nil 不能传给 Registry，否则会去连一个不存在的 Consul。
	if reg != nil {
		opts = append(opts, micro.Registry(reg))
	}
	// NewService 只组装对象，此时还没有开始监听。
	rpcService := micro.NewService(opts...)

	// 先连上 MySQL 服务器，再执行 CREATE DATABASE IF NOT EXISTS。
	if err := common.EnsureDatabase(mysqlCfg); err != nil {
		log.Fatal("create database err :", err, "请检查 user-service/config/mysql-user.json 的账号密码")
	}
	// 函数名带 Consul，是课程原来的名字。现在传进来的也可以是本地 JSON。
	db, err := common.GetMysqlFromConsul(mysqlCfg)
	if err != nil {
		log.Fatal("mysql init err :", err)
	}
	// AutoMigrate 按 User 结构体建表或补列，不会删除已有列，也不会改已有数据。
	if err := common.AutoMigrateCommented(db, &model.User{}, &model.UserAddress{}); err != nil {
		log.Fatal("migrate user table err :", err)
	}
	// 登录令牌存在 Redis 里。连不上就退出，避免服务看起来启动了、登录时才失败。
	red, err := common.GetRedisFromConsul(redisCfg)
	if err != nil {
		log.Fatal("redis init err :", err, "请确认本机 Redis 已启动，地址见 config/redis.json")
	}

	// 从内往外组装：数据库和 Redis → repository → service → handler。
	userDataService := service.NewUserDataService(repository.NewUserRepository(db, red))
	userHandler := &handler.User{UserDataService: userDataService}
	// 同一个结构体注册两次。登录和校验令牌是 proto 里两个独立的 service。
	proto.RegisterLoginHandler(rpcService.Server(), userHandler)
	proto.RegisterGetUserTokenHandler(rpcService.Server(), userHandler)
	log.Println("shop-user listening :8081")
	// Run 开始监听并堵住当前协程。写在它后面的代码，只有服务停掉才会执行。
	if err := rpcService.Run(); err != nil {
		log.Println("start user service err :", err)
	}
}

// loadConfig 优先读 Consul。配置中心不在线或键是空的，就用编译进程序的本地 JSON。
// 命名返回值 mysqlCfg、redisCfg、reg 在函数里可以直接赋值，return 时按这个顺序送出。
func loadConfig() (mysqlCfg, redisCfg *viper.Viper, reg registry.Registry) {
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-user.json"), mustLocal("config/redis.json"), nil
	}

	// UserFileKey 的值是 "mysql-user"，RedisFileKey 的值是 "redis"。
	mysqlCfg, _ = common.GetConsulConfig(common.ConsulStr, common.UserFileKey)
	if mysqlCfg == nil || mysqlCfg.GetString("host") == "" {
		log.Println("consul mysql config empty, use local config")
		mysqlCfg = mustLocal("config/mysql-user.json")
	}
	redisCfg, _ = common.GetConsulConfig(common.ConsulStr, common.RedisFileKey)
	if redisCfg == nil || redisCfg.GetString("addr") == "" {
		log.Println("consul redis config empty, use local config")
		redisCfg = mustLocal("config/redis.json")
	}
	// 这个函数由 go-micro 调用，options 是它传进来的。我们只填注册中心地址。
	reg = consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{common.ConsulReistStr}
	})
	return mysqlCfg, redisCfg, reg
}

// mustLocal 从编译进程序的文件里读 JSON。名字里的 must 表示失败就退出，不返回错误。
func mustLocal(name string) *viper.Viper {
	body, err := localConfig.ReadFile(name)
	if err != nil {
		log.Fatal("read local config err :", name, err)
	}
	conf, err := common.ParseJSON(body)
	if err != nil {
		log.Fatal("parse local config err :", name, err)
	}
	log.Println("loaded local config :", name)
	return conf
}
