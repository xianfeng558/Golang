// 命令：在 trade-order 目录执行 go run .
// 作用：启动名为 trade-order 的 RPC 服务，监听 8085。
// 同时在 9092 端口暴露 Prometheus 指标。HTTP 网关在 client 目录，端口 6669。
package main

import (
	"embed"
	"log"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	"github.com/go-micro/plugins/v4/wrapper/monitoring/prometheus"
	ratelimiter "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"trade-order/domain/model"
	"trade-order/domain/repository"
	"trade-order/domain/service"
	"trade-order/handler"
	"trade-order/proto"
)

//go:embed config/mysql-trade.json
var localConfig embed.FS

func main() {
	mysqlCfg, reg := loadConfig()
	if closer := common.UseTracer("trade-order"); closer != nil {
		defer closer.Close()
	}
	common.PrometheusBoot(9092)

	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),
		micro.RegisterInterval(time.Second * 30),
		micro.Name("trade-order"),
		micro.Address(":8085"),
		micro.Version("v1"),
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())),
		micro.WrapHandler(ratelimiter.NewHandlerWrapper(common.QPS)),
		micro.WrapHandler(prometheus.NewHandlerWrapper()),
	}
	if reg != nil {
		opts = append(opts, micro.Registry(reg))
	}
	rpcService := micro.NewService(opts...)

	if err := common.EnsureDatabase(mysqlCfg); err != nil {
		log.Fatal("create database err :", err, "请检查 trade-order/config/mysql-trade.json")
	}
	db, err := common.GetMysqlFromConsul(mysqlCfg)
	if err != nil {
		log.Fatal("mysql init err :", err)
	}
	if err := db.AutoMigrate(&model.TraderOrder{}); err != nil {
		log.Fatal("migrate trade table err :", err)
	}

	tradeService := service.NewTradeOrderService(repository.NewTradeRepository(db))
	h := &handler.TradeOrderHandler{TradeOrderService: tradeService}
	proto.RegisterAddTradeOrderHandler(rpcService.Server(), h)
	proto.RegisterUpdateTradeOrderHandler(rpcService.Server(), h)
	proto.RegisterFindOrderHandler(rpcService.Server(), h)
	log.Println("trade-order listening :8085")
	if err := rpcService.Run(); err != nil {
		log.Println("start trade service err :", err)
	}
}

func loadConfig() (*viper.Viper, registry.Registry) {
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-trade.json"), nil
	}
	conf, _ := common.GetConsulConfig(common.ConsulStr, common.TradeFileKey)
	if conf == nil || conf.GetString("host") == "" {
		log.Println("consul trade config empty, use local config")
		conf = mustLocal("config/mysql-trade.json")
	}
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{common.ConsulReistStr}
	})
	return conf, reg
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
