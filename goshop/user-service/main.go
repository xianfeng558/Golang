// 命令：在 user-service 目录执行 go run .
// 作用：启动名为 shop-user 的 RPC 服务，监听 8081。
// 浏览器不直接访问 8081，而是访问 client 目录里 8088 端口的 HTTP 网关。
package main

import (
	"embed"
	"log"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"gouser/domain/model"
	"gouser/domain/repository"
	"gouser/domain/service"
	"gouser/handler"
	"gouser/proto"
)

// localConfig 在编译时把 JSON 打进可执行文件。换机器不用额外拷贝配置，改完要重新 go run。
//
//go:embed config/mysql-user.json config/redis.json
var localConfig embed.FS

func main() {
	mysqlCfg, redisCfg, reg := loadConfig()

	// micro.Option 是启动参数。Name 是服务发现用的名字，其它服务靠这个名字来找它。
	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),      // 注册信息 30 秒过期，进程挂了会被摘掉
		micro.RegisterInterval(time.Second * 30), // 每 30 秒续一次注册
		micro.Name("shop-user"),
		micro.Address(":8081"),
		micro.Version("v1"),
	}
	if reg != nil {
		opts = append(opts, micro.Registry(reg))
	}
	rpcService := micro.NewService(opts...)

	if err := common.EnsureDatabase(mysqlCfg); err != nil {
		log.Fatal("create database err :", err, "请检查 user-service/config/mysql-user.json 的账号密码")
	}
	db, err := common.GetMysqlFromConsul(mysqlCfg)
	if err != nil {
		log.Fatal("mysql init err :", err)
	}
	// AutoMigrate 根据结构体建表或补列，不会删除已有列。
	if err := db.AutoMigrate(&model.User{}); err != nil {
		log.Fatal("migrate user table err :", err)
	}
	red, err := common.GetRedisFromConsul(redisCfg)
	if err != nil {
		log.Fatal("redis init err :", err, "请确认本机 Redis 已启动，地址见 config/redis.json")
	}

	userDataService := service.NewUserDataService(repository.NewUserRepository(db, red))
	userHandler := &handler.User{UserDataService: userDataService}
	// proto 生成的 Register 函数把方法挂到 RPC 服务器上。同一个结构体可以实现多个 service。
	proto.RegisterLoginHandler(rpcService.Server(), userHandler)
	proto.RegisterGetUserTokenHandler(rpcService.Server(), userHandler)
	log.Println("shop-user listening :8081")
	if err := rpcService.Run(); err != nil {
		log.Println("start user service err :", err)
	}
}

// loadConfig 优先读 Consul。配置中心不在线或键是空的，就用编译进程序的本地 JSON。
// 第三个返回值是注册中心。没有 Consul 时返回 nil，go-micro 会退回到 mDNS，在本机互相发现。
func loadConfig() (mysqlCfg, redisCfg *viper.Viper, reg registry.Registry) {
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-user.json"), mustLocal("config/redis.json"), nil
	}

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
	reg = consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{common.ConsulReistStr}
	})
	return mysqlCfg, redisCfg, reg
}

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
