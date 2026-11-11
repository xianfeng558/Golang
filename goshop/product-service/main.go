// 命令：在 product-service 目录执行 go run .
// 作用：启动名为 shop-product 的 RPC 服务，监听 8082。
// HTTP 网关在 client 目录，端口 8087。第一次启动会建库、建表，并写入一条示例商品。
package main

import (
	"embed"
	"log"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing" // 给每个 RPC 包一层耗时追踪
	_ "github.com/go-sql-driver/mysql"                                      // 只执行初始化，让 GORM 能连 MySQL
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"goproduct/domain/model"
	"goproduct/domain/repository"
	"goproduct/domain/service"
	"goproduct/handler"
	"goproduct/proto"
	"gorm.io/gorm"        // seedProduct 的参数类型。*gorm.DB 是数据库连接
	"gorm.io/gorm/clause" // 分类种子按主键忽略重复，重复启动不会插出第二份
)

// 编译时把商品库的 JSON 收进程序。改完要重新 go run。
//
//go:embed config/mysql-product.json
var localConfig embed.FS

// main 的顺序：读配置、开追踪、组装 RPC、建库建表、写入示例商品、注册五个商品方法。
func main() {
	consulConfig, reg := loadConfig()
	// UseTracer 连上 Jaeger 时 closer 才有值。defer 等 main 退出时才关闭。
	if closer := common.UseTracer("shop-product"); closer != nil {
		defer closer.Close()
	}

	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),      // 注册信息 30 秒过期
		micro.RegisterInterval(time.Second * 30), // 每 30 秒续一次注册
		micro.Name("shop-product"),               // 购物车网关按这个名字找商品服务
		micro.Address(":8082"),
		micro.Version("v1"),
		// WrapHandler 像一层洋葱皮：请求先过追踪，再进真正的商品方法。
		micro.WrapHandler(opentracing2.NewHandlerWrapper(opentracing.GlobalTracer())),
	}
	if reg != nil {
		opts = append(opts, micro.Registry(reg))
	}
	rpcService := micro.NewService(opts...)

	if err := common.EnsureDatabase(consulConfig); err != nil {
		log.Fatal("create database err :", err, "请检查 product-service/config/mysql-product.json")
	}
	db, err := common.GetMysqlFromConsul(consulConfig)
	if err != nil {
		log.Fatal("mysql init err :", err)
	}
	// 一次迁移四张表：商品、规格、图文详情、图片。
	if err := db.AutoMigrate(&model.Category{}, &model.Product{}, &model.ProductSku{}, &model.ProductDetailRow{}, &model.ProductPicture{}); err != nil {
		log.Fatal("migrate product table err :", err)
	}
	seedCategory(db)
	seedProduct(db)

	productDataService := service.NewProductDataService(repository.NewProductRepository(db))
	h := &handler.ProductHandler{ProductDataService: productDataService}
	// 五个 RPC 共用同一个 handler。proto 把它们分成了五个 service。
	proto.RegisterPageHandler(rpcService.Server(), h)
	proto.RegisterShowProductDetailHandler(rpcService.Server(), h)
	proto.RegisterShowProductSkuHandler(rpcService.Server(), h)
	proto.RegisterShowDetailSkuHandler(rpcService.Server(), h)
	proto.RegisterUpdateSkuHandler(rpcService.Server(), h)
	log.Println("shop-product listening :8082")
	if err := rpcService.Run(); err != nil {
		log.Println("start product service err :", err)
	}
}

// loadConfig 准备 MySQL 配置和可选的 Consul。第二个返回值可以是 nil。
func loadConfig() (*viper.Viper, registry.Registry) {
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-product.json"), nil
	}
	// ProductFileKey 的值是 "mysql-product"。
	conf, _ := common.GetConsulConfig(common.ConsulStr, common.ProductFileKey)
	if conf == nil || conf.GetString("host") == "" {
		log.Println("consul mysql config empty, use local config")
		conf = mustLocal("config/mysql-product.json")
	}
	reg := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{common.ConsulReistStr}
	})
	return conf, reg
}

// mustLocal 从编译进程序的文件里读 JSON。读失败就退出进程。
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

// seedCategory 保证示例商品用到的分类 31 存在，并补上两个顶级分类。
// 主键冲突时什么都不做，所以重复启动不会插出重名行。
func seedCategory(db *gorm.DB) {
	rows := []model.Category{
		{ID: 1, ParentID: 0, Name: "数码", Sort: 1, IsEnable: true},
		{ID: 2, ParentID: 0, Name: "图书", Sort: 2, IsEnable: true},
		{ID: 31, ParentID: 2, Name: "学习示例", Sort: 1, IsEnable: true},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error; err != nil {
		log.Println("seed category err :", err)
	}
}

// seedProduct 在空表里放一条商品、一条详情、一张图、一个库存为 100 的规格。
// 购物车加购时会读这些数据。表里已经有商品时不会再插第二条。
func seedProduct(db *gorm.DB) {
	var product model.Product
	var count int64
	// Count 把行数写进 count。count 前面的 & 表示把地址交给 GORM，让它填这个变量。
	db.Model(&model.Product{}).Count(&count)
	if count == 0 {
		now := time.Now()
		product = model.Product{
			Name:              "Go 微服务入门示例商品",
			CategoryId:        31,
			StartingPrice:     98,
			MainPicture:       "https://msb-edu-prod.oss-cn-beijing.aliyuncs.com/mall-product/product/8c23730b-d86c-43ed-9db2-9efaa1028115.png",
			RemoteAreaPostage: 10,
			IsEnable:          true,
			Remark:            "本地种子数据",
			CreateTime:        now,
			UpdateTime:        now,
		}
		// Create 成功后，GORM 会把数据库生成的主键写回 product.ID。
		if err := db.Create(&product).Error; err != nil {
			log.Println("seed product err :", err)
			return
		}
		db.Create(&model.ProductDetailRow{ProductId: product.ID, Detail: "<p>这是自动写入的示例详情，用来打通商品详情接口。</p>"})
		db.Create(&model.ProductPicture{ProductId: product.ID, Picture: product.MainPicture})
		log.Println("seeded product id", product.ID)
	} else {
		// 表里已有商品时，取出第一行，后面用它的 ID 检查详情和图片是否缺失。
		db.First(&product)
	}
	if product.ID != 0 {
		var detailCount int64
		db.Model(&model.ProductDetailRow{}).Where("product_id = ?", product.ID).Count(&detailCount)
		if detailCount == 0 {
			db.Create(&model.ProductDetailRow{ProductId: product.ID, Detail: "<p>这是自动写入的示例详情，用来打通商品详情接口。</p>"})
		}
		var pictureCount int64
		db.Model(&model.ProductPicture{}).Where("product_id = ?", product.ID).Count(&pictureCount)
		if pictureCount == 0 {
			db.Create(&model.ProductPicture{ProductId: product.ID, Picture: product.MainPicture})
		}
	}

	var skuCount int64
	db.Model(&model.ProductSku{}).Count(&skuCount)
	if skuCount == 0 && product.ID != 0 {
		db.Create(&model.ProductSku{
			ProductId:           product.ID,
			Name:                "默认规格",
			AttributeSymbolList: "默认",
			SellPrice:           product.StartingPrice,
			Stock:               100,
		})
		log.Println("seeded sku for product", product.ID)
	}
}
