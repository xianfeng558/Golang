// 命令：在 trade-order 目录执行 go run .
// 作用：启动名为 trade-order 的 RPC 服务，监听 8085。
// 同时在 9092 端口暴露 Prometheus 指标。浏览器不直接访问 8085，HTTP 网关在 client 目录，端口 6669。
package main

import (
	"embed" // 编译时把 config 目录里的 JSON 打进可执行文件
	"log"   // 打印启动日志。Fatal 会打印后让进程退出
	"time"  // 注册中心的心跳间隔用到秒

	// common 是五个服务共用的工具：配置、MySQL、追踪、监控。
	// 引号里是 go.mod 的模块路径，go.work 会把它指到本地的 goshop/common。
	common "git.mashibing.com/msb_47094/shopping-comm"
	// consul 是注册中心插件。本机 8500 没开时，下面不会使用它。
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	// 三个 wrapper 像洋葱皮，套在真正的业务函数外面：
	// prometheus 记录调用次数，ratelimiter 限制每秒请求数，opentracing 记录耗时。
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	ratelimiter "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	// 前面的下划线表示只为了执行这个包的初始化，本文件不直接调用它。
	// GORM 连 MySQL 时需要这个驱动先把自己注册好。
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentracing/opentracing-go" // 链路追踪的接口。没开 Jaeger 时也不影响下单
	"github.com/spf13/viper"                // 读 JSON 配置，后面用 GetString("host") 取字段
	"go-micro.dev/v4"                       // RPC 框架。NewService、Name、Address 都来自这里
	"go-micro.dev/v4/registry"              // 注册中心的参数类型，例如地址列表
	"trade-order/domain/model"              // 订单表结构，AutoMigrate 靠它建表
	"trade-order/domain/repository"         // 写 SQL 的那一层
	"trade-order/domain/service"            // 把 repository 交给 handler 的薄薄一层
	"trade-order/handler"                   // 创建订单、改状态、查询的 RPC 实现
	"trade-order/proto"                     // 由 trade.proto 生成的注册函数
)

// go:embed 必须紧挨着下面这个变量，中间不能插别的语句。
// 编译时把 config/mysql-trade.json 收进程序。改完 JSON 要重新 go run，只改文件不重启不会生效。
//
//go:embed config/mysql-trade.json
var localConfig embed.FS

// main 是进程入口。Go 规定可执行程序必须有 package main 和 func main。
// 顺序是：读配置、开追踪和监控、组装 RPC、连数据库、注册三个订单方法，最后 Run 堵住不返回。
func main() {
	// loadConfig 返回两样东西：MySQL 配置，以及可选的 Consul。
	// 没有 Consul 时 reg 是 nil，服务改用本机 mDNS 让别的进程找到自己。
	mysqlCfg, reg := loadConfig()

	// UseTracer 连接 Jaeger。closer 非空说明连上了，函数结束时用 defer 关掉。
	// defer 不立刻执行，而是等 main 退出时才执行。
	if closer := common.UseTracer("trade-order"); closer != nil {
		defer closer.Close()
	}
	// 在 9092 打开指标页，浏览器访问 http://127.0.0.1:9092/metrics。
	// 这个调用放在协程里，不会挡住后面的启动。
	common.PrometheusBoot(9092)

	// micro.Option 是一条启动参数。切片用来收集多条，有 Consul 时再 append 追加。
	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),      // 注册信息 30 秒过期，防止进程死后地址还留着
		micro.RegisterInterval(time.Second * 30), // 每 30 秒续一次注册
		micro.Name("trade-order"),                // 别的服务靠这个名字来调用，不写死 IP
		micro.Address(":8085"),                   // RPC 监听本机 8085。冒号表示所有网卡
		micro.Version("v1"),                      // 版本号，给注册中心看，业务判断不用它
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())), // 给每个 RPC 包一层追踪
		micro.WrapHandler(ratelimiter.NewHandlerWrapper(common.QPS)),                  // 每秒大约放行 100 个请求
		micro.WrapHandler(prometheus.NewHandlerWrapper()),                             // 统计每个 RPC 的次数和耗时
	}
	// reg 有值才加入注册中心。nil 不能传给 Registry，否则会去连一个不存在的 Consul。
	if reg != nil {
		opts = append(opts, micro.Registry(reg))
	}
	// NewService 只是把上面的参数组装成服务对象，此时还没有开始监听。
	rpcService := micro.NewService(opts...)

	// EnsureDatabase 先连上 MySQL 服务器，再执行 CREATE DATABASE IF NOT EXISTS。
	// log.Fatal 打印错误并结束进程。库都建不出来时，继续往下跑没有意义。
	if err := common.EnsureDatabase(mysqlCfg); err != nil {
		log.Fatal("create database err :", err, "请检查 trade-order/config/mysql-trade.json")
	}
	// GetMysqlFromConsul 这个名字来自课程。现在传进来的也可以是本地 JSON 解析出的配置。
	db, err := common.GetMysqlFromConsul(mysqlCfg)
	if err != nil {
		log.Fatal("mysql init err :", err)
	}
	// AutoMigrate 按 TraderOrder 结构体建表、补列。它不会删列，也不会改已有数据。
	if err := db.AutoMigrate(&model.TraderOrder{}); err != nil {
		log.Fatal("migrate trade table err :", err)
	}

	// 从内往外组装：数据库连接 → repository → service → handler。
	// & 取地址。handler 的字段是接口，所以这里放的是 New 函数返回的对象。
	tradeService := service.NewTradeOrderService(repository.NewTradeRepository(db))
	h := &handler.TradeOrderHandler{TradeOrderService: tradeService}
	// 同一个 h 注册三次，因为创建、修改、查询是 proto 里三个独立的 service。
	proto.RegisterAddTradeOrderHandler(rpcService.Server(), h)
	proto.RegisterUpdateTradeOrderHandler(rpcService.Server(), h)
	proto.RegisterFindOrderHandler(rpcService.Server(), h)
	log.Println("trade-order listening :8085")
	// Run 开始监听并堵住当前协程。写在它后面的代码，只有服务停掉才会执行。
	if err := rpcService.Run(); err != nil {
		log.Println("start trade service err :", err)
	}
}

// loadConfig 准备数据库配置和注册中心。
// 返回值有两个：*viper.Viper 是配置，registry.Registry 是注册中心。第二个可以是 nil。
func loadConfig() (*viper.Viper, registry.Registry) {
	// ConsulReachable 用 1.5 秒试连 127.0.0.1:8500。连不上就只用本地 JSON。
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-trade.json"), nil
	}
	// TradeFileKey 的值是 "mysql-trade"，也就是 Consul 配置中心里的键名。
	conf, _ := common.GetConsulConfig(common.ConsulStr, common.TradeFileKey)
	// host 为空说明远程配置没读到内容，退回本地文件。
	if conf == nil || conf.GetString("host") == "" {
		log.Println("consul trade config empty, use local config")
		conf = mustLocal("config/mysql-trade.json")
	}
	// NewRegistry 的参数是一个函数，go-micro 会把 options 传进来，我们只填地址。
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{common.ConsulReistStr}
	})
	return conf, reg
}

// mustLocal 从编译进程序的文件里读 JSON。名字里的 must 表示失败就退出，不返回错误。
func mustLocal(name string) *viper.Viper {
	// ReadFile 的参数是 embed 时的路径，不是运行时的工作目录。
	body, err := localConfig.ReadFile(name)
	if err != nil {
		log.Fatal("read local config err :", name, err)
	}
	// ParseJSON 把字节变成 viper，后面才能用 GetString("user") 这类方法。
	conf, err := common.ParseJSON(body)
	if err != nil {
		log.Fatal("parse local config err :", name, err)
	}
	log.Println("loaded local config :", name)
	return conf
}
