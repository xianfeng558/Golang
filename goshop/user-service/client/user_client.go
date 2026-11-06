// 用户 HTTP 网关，端口 8088。
// 它只做两件事：用 Gin 接收浏览器的查询参数，再调用 8081 上的 shop-user。
// 先启动 user-service，再启动本程序，否则登录会报找不到服务。
package main

import (
	"context"
	"log"
	"strconv"

	common "git.mashibing.com/msb_47094/shopping-comm"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	"github.com/gin-gonic/gin"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/web"
	"gouser/proto"
)

// getClient 每次登录都新建一个 RPC 客户端。
// 返回类型是 proto.LoginService，这是 .proto 生成的接口，里面有 Login 方法。
func getClient() proto.LoginService {
	// micro.Option 是启动选项。切片从空开始，有 Consul 再追加注册中心。
	opts := []micro.Option{}
	if common.ConsulReachable() {
		opts = append(opts, micro.Registry(consul.NewRegistry(func(options *registry.Options) {
			options.Addrs = []string{common.ConsulReistStr}
		})))
	} else {
		log.Println("consul unreachable, discover shop-user via mdns")
	}
	rpcServer := micro.NewService(opts...)
	return proto.NewLoginService("shop-user", rpcServer.Client())
}

func main() {
	router := gin.Default()

	router.Handle("GET", "toLogin", func(context *gin.Context) {
		context.String(200, "to Loging ....")
	})

	// 例子：/login?phone=13800000000&clientId=0&systemId=0&verificationCode=6666
	router.GET("/login", func(c *gin.Context) {
		client := getClient()
		// 查询参数全是字符串。Atoi 转整数失败时返回 0 和错误，这里忽略错误，非法数字会变成 0。
		clientId, _ := strconv.Atoi(c.Request.FormValue("clientId"))
		phone := c.Request.FormValue("phone")
		systemId, _ := strconv.Atoi(c.Request.FormValue("systemId"))
		verificationCode := c.Request.FormValue("verificationCode")
		// 复合字面量：按字段名填 proto 结构体。没写到的字段是零值，数字是 0，字符串是 ""。
		req := &proto.LoginRequest{
			ClientId:         int32(clientId),
			Phone:            phone,
			SystemId:         int32(systemId),
			VerificationCode: verificationCode,
		}
		resp, err := client.Login(context.TODO(), req)
		if err != nil {
			log.Println(err.Error())
			common.RespFail(c.Writer, resp, "登录失败")
			return
		}
		common.RespOK(c.Writer, resp, "登录成功")
	})

	// web.NewService 把 Gin 挂到 8088。Run 会一直阻塞，所以它是 main 的最后一行。
	service := web.NewService(
		web.Address(":8088"),
		web.Handler(router),
	)
	service.Run()
}
