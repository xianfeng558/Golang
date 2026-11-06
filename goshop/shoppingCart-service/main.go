// 命令：在 shoppingCart-service 目录执行 go run .
// 作用：启动名为 shop-cart 的 RPC 服务，监听 8084。
// 购物车表建在 product_center 库里，因为算总价的 SQL 要和 product_sku 做 JOIN。
package main

import (
	"embed"
	"log"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	ratelimiter "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"shoppingCart/domain/model"
	"shoppingCart/domain/repository"
	"shoppingCart/domain/service"
	"shoppingCart/handler"
	"shoppingCart/proto"
)

//go:embed config/mysql-cart.json
var localConfig embed.FS

func main() {
	mysqlCfg, reg := loadConfig()
	if closer := common.UseTracer("shop-cart"); closer != nil {
		defer closer.Close()
	}

	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),
		micro.RegisterInterval(time.Second * 30),
		micro.Name("shop-cart"),
		micro.Address(":8084"),
		micro.Version("v1"),
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())),
		micro.WrapHandler(ratelimiter.NewHandlerWrapper(common.QPS)),
	}
	if reg != nil {
		opts = append(opts, micro.Registry(reg))
	}
	rpcService := micro.NewService(opts...)

	if err := common.EnsureDatabase(mysqlCfg); err != nil {
		log.Fatal("create database err :", err, "请检查 shoppingCart-service/config/mysql-cart.json")
	}
	db, err := common.GetMysqlFromConsul(mysqlCfg)
	if err != nil {
		log.Fatal("mysql init err :", err)
	}
	if err := db.AutoMigrate(&model.ShoppingCart{}); err != nil {
		log.Fatal("migrate cart table err :", err)
	}

	cartService := service.NewCartService(repository.NewCartRepository(db))
	h := &handler.CartHandler{CartService: cartService}
	proto.RegisterAddCartHandler(rpcService.Server(), h)
	proto.RegisterUpdateCartHandler(rpcService.Server(), h)
	proto.RegisterGetOrderTotalHandler(rpcService.Server(), h)
	proto.RegisterFindCartHandler(rpcService.Server(), h)
	log.Println("shop-cart listening :8084")
	if err := rpcService.Run(); err != nil {
		log.Println("start cart service err :", err)
	}
}

func loadConfig() (*viper.Viper, registry.Registry) {
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-cart.json"), nil
	}
	conf, _ := common.GetConsulConfig(common.ConsulStr, "mysql-cart")
	if conf == nil || conf.GetString("host") == "" {
		log.Println("consul cart config empty, use local config")
		conf = mustLocal("config/mysql-cart.json")
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
