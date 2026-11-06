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
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	_ "github.com/go-sql-driver/mysql"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"goproduct/domain/model"
	"goproduct/domain/repository"
	"goproduct/domain/service"
	"goproduct/handler"
	"goproduct/proto"
	"gorm.io/gorm"
)

//go:embed config/mysql-product.json
var localConfig embed.FS

func main() {
	consulConfig, reg := loadConfig()
	if closer := common.UseTracer("shop-product"); closer != nil {
		defer closer.Close()
	}

	opts := []micro.Option{
		micro.RegisterTTL(time.Second * 30),
		micro.RegisterInterval(time.Second * 30),
		micro.Name("shop-product"),
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
	if err := db.AutoMigrate(&model.Product{}, &model.ProductSku{}, &model.ProductDetailRow{}, &model.ProductPicture{}); err != nil {
		log.Fatal("migrate product table err :", err)
	}
	seedProduct(db)

	productDataService := service.NewProductDataService(repository.NewProductRepository(db))
	h := &handler.ProductHandler{ProductDataService: productDataService}
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

func loadConfig() (*viper.Viper, registry.Registry) {
	if !common.ConsulReachable() {
		log.Println("consul unreachable, use local config :", common.ConsulReistStr)
		return mustLocal("config/mysql-product.json"), nil
	}
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

// seedProduct 在空表里放一条商品、一条详情、一张图、一个库存为 100 的规格。
// 购物车加购时会读这些数据。已经有数据时不会重复插入。
func seedProduct(db *gorm.DB) {
	var product model.Product
	var count int64
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
		if err := db.Create(&product).Error; err != nil {
			log.Println("seed product err :", err)
			return
		}
		db.Create(&model.ProductDetailRow{ProductId: product.ID, Detail: "<p>这是自动写入的示例详情，用来打通商品详情接口。</p>"})
		db.Create(&model.ProductPicture{ProductId: product.ID, Picture: product.MainPicture})
		log.Println("seeded product id", product.ID)
	} else {
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
