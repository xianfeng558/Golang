// 这是商品的 HTTP 网关，自己不查数据库。
// 浏览器访问 8087，本程序再调用 8082 上名为 shop-product 的 RPC 服务。
// 所以要先在另一个终端执行 product-service 目录下的 go run .，再启动本文件。
package main

import (
	"context"
	"log"
	"net"
	"strconv"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	"github.com/opentracing/opentracing-go"
	"go-micro.dev/v4/web"
	"goproduct/proto"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
)

// main 是程序入口。Go 规定可执行程序必须有 package main 和 func main。
func main() {
	// gin.Default 自带访问日志和宕机恢复。router 上每注册一个路径，就多一个 HTTP 接口。
	router := gin.Default()

	// Handle 的前两个参数是方法和路径。第三个是处理函数，请求进来时 Gin 会调用它。
	// 这里的 context 是参数名，和标准库的 context 包不是一回事，只是个 *gin.Context。
	router.Handle("GET", "toPage", func(context *gin.Context) {
		context.String(200, "to toPage ....")
	})

	// 链路追踪失败时服务仍继续。defer 要写在 err == nil 的分支里，避免对空的 closer 调用 Close。
	t, closer, err := common.NewTracer("shop-product-client", common.ConsulIp+":6831")
	if err != nil {
		log.Println("tracer err :", err)
	} else {
		defer closer.Close()
		opentracing.SetGlobalTracer(t)
	}

	// []micro.Option 是「启动参数」切片。后面可以 append 追加，所以 Consul 在不在都能组装。
	opts := []micro.Option{
		micro.WrapClient(opentracing2.NewClientWrapper(opentracing.GlobalTracer())),
	}
	webOpts := []web.Option{
		web.Address(":8087"),
		web.Name("shop-product-client"),
		web.Handler(router),
	}
	if consulReachable(common.ConsulReistStr) {
		reg := consul.NewRegistry(func(options *registry.Options) {
			options.Addrs = []string{common.ConsulReistStr}
		})
		opts = append(opts, micro.Registry(reg))
		webOpts = append(webOpts, web.Registry(reg))
	} else {
		log.Println("consul unreachable, discover shop-product via mdns")
	}
	// NewService 得到的是 RPC 客户端的外壳，不是又开了一个商品数据库服务。
	rpcServer := micro.NewService(opts...)
	// "shop-product" 是对方注册时用的名字。go-micro 靠这个名字找到 8082 上的进程。
	client := proto.NewPageService("shop-product", rpcServer.Client())

	// /page?length=10&pageIndex=1 查询商品列表。
	router.GET("/page", func(c *gin.Context) {
		// FormValue 读到的是字符串。Atoi 把 "10" 转成整数，第二个返回值是错误，这里用 _ 丢掉。
		length, _ := strconv.Atoi(c.Request.FormValue("length"))
		pageIndex, _ := strconv.Atoi(c.Request.FormValue("pageIndex"))
		if length == 0 {
			length = 15
		}
		if pageIndex == 0 {
			pageIndex = 1
		}
		// & 取地址。RPC 参数约定传指针，这样函数里改的是同一块内存，不用来回复制整个结构体。
		// int32 是协议里定的 32 位整数，和 Atoi 返回的 int 宽度可能不同，所以要显式转换。
		req := &proto.PageReq{
			Length:    int32(length),
			PageIndex: int32(pageIndex),
		}
		// context.TODO 表示「暂时不设置超时」。正式项目应使用带截止时间的 context。
		resp, err := client.Page(context.TODO(), req)
		log.Println(" /page  :", resp)
		if err != nil {
			log.Println(err.Error())
			common.RespFail(c.Writer, resp, "请求失败")
			return
		}
		common.RespListOK(c.Writer, resp, "请求成功", resp.Rows, resp.Total, "请求成功")
	})

	// 同一个 RPC 进程可以挂多个 service。详情和列表是 product.proto 里分开声明的。
	clientA := proto.NewShowProductDetailService("shop-product", rpcServer.Client())
	router.GET("/showProductDetail", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Request.FormValue("id"))
		req := &proto.ProductDetailReq{
			Id: int32(id),
		}
		resp, err := clientA.ShowProductDetail(context.TODO(), req)
		log.Println(" /showProductDetail  :", resp)
		if err != nil {
			log.Println(err.Error())
			common.RespFail(c.Writer, resp, "请求失败")
			return
		}
		common.RespOK(c.Writer, resp, "请求成功")
	})

	clientSKU := proto.NewShowProductSkuService("shop-product", rpcServer.Client())
	router.GET("/sku", func(c *gin.Context) {
		id, _ := strconv.Atoi(c.Request.FormValue("productId"))
		req := &proto.ProductSkuReq{
			ProductId: int32(id),
		}
		resp, err := clientSKU.ShowProductSku(context.TODO(), req)
		log.Println(" /sku  :", resp)
		if err != nil {
			log.Println(err.Error())
			common.RespFail(c.Writer, resp, "请求失败")
			return
		}
		common.RespListOK(c.Writer, resp, "请求成功", 0, 0, "请求成功")
	})

	// Run 会堵住当前协程，直到进程退出。写在它后面的代码不会执行。
	service := web.NewService(webOpts...)
	service.Run()
}

// consulReachable 在 1.5 秒内连接注册中心。连上说明 Consul 开着，连不上就改用 mDNS。
// DialTimeout 的返回值 conn 是一条 TCP 连接，探测完必须 Close，否则会占着端口。
func consulReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 1500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
