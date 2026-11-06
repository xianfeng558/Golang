// 订单 HTTP 网关，端口 6669。
// 下单接口会调用购物车和用户服务，并可选地交给 DTM 做分布式事务。
// 先启动 user-service、shoppingCart-service、trade-order 这三个 RPC，再启动本程序。
package main

import (
	"context"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/go-micro/plugins/v4/wrapper/select/roundrobin"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	"github.com/lithammer/shortuuid/v3"
	"github.com/opentracing/opentracing-go"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/web"
	"strconv"
	//"goproduct/common"
	common "git.mashibing.com/msb_47094/shopping-comm"
	"log"
	"net"
	"net/http"
	"trade-order/proto"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
)

// 获取远程服务的客户端
func main() {
	resp := &proto.AddTradeOrderResp{}
	router := gin.Default()
	var consulReg registry.Registry
	if common.ConsulReachable() {
		consulReg = consul.NewRegistry(func(options *registry.Options) {
			options.Addrs = []string{common.ConsulReistStr}
		})
	} else {
		log.Println("consul unreachable, discover services via mdns")
	}
	if closer := common.UseTracer("trade-order-client"); closer != nil {
		defer closer.Close()
	}

	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go func() {
		err := http.ListenAndServe(net.JoinHostPort("127.0.0.1", "9097"), hystrixStreamHandler)
		if err != nil {
			log.Println("hystrix stream :", err)
		}
	}()

	opts := []micro.Option{
		micro.WrapClient(opentracing2.NewClientWrapper(opentracing.GlobalTracer())),
		micro.WrapClient(NewClientHystrixWrapper()),
		micro.WrapClient(roundrobin.NewClientWrapper()),
	}
	webOpts := []web.Option{
		web.Address(":6669"),
		web.Name("trade-order-client"),
		web.Handler(router),
	}
	if consulReg != nil {
		opts = append(opts, micro.Registry(consulReg))
		webOpts = append(webOpts, web.Registry(consulReg))
	}
	rpcServer := micro.NewService(opts...)
	//AddCartClient := proto.NewAddCartService("shop-cart", rpcServer.Client())
	UpdateCartClient := proto.NewUpdateCartService("shop-cart", rpcServer.Client())
	//ShowProductDetailClient := proto.NewShowProductDetailService("shop-product", rpcServer.Client())
	//ShowDetailSkuClient := proto.NewShowDetailSkuService("shop-product", rpcServer.Client())
	GetUserTokenClient := proto.NewGetUserTokenService("shop-user", rpcServer.Client())
	GetOrderTotalClient := proto.NewGetOrderTotalService("shop-cart", rpcServer.Client())
	AddTraderClient := proto.NewAddTradeOrderService("trade-order", rpcServer.Client())
	UpdateTraderClient := proto.NewUpdateTradeOrderService("trade-order", rpcServer.Client())
	FindCartClient := proto.NewFindCartService("shop-cart", rpcServer.Client())
	FindOrderClient := proto.NewFindOrderService("trade-order", rpcServer.Client())

	//开始拆分 DTM服务

	router.POST("/updateCart", func(c *gin.Context) {
		req := &proto.UpdateCartReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		req.IsDeleted = true
		_, err := UpdateCartClient.UpdateCart(context.TODO(), req)
		if err != nil {
			log.Println("/updateCart err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "删除购物车失败!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"updateCart": "SUCCESS", "Message": "删除购物车成功!"})
	})
	router.POST("/updateCart-compensate", func(c *gin.Context) {
		req := &proto.UpdateCartReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		req.IsDeleted = false
		_, err := UpdateCartClient.UpdateCart(context.TODO(), req)
		if err != nil {
			log.Println("/updateCart err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "回滚购物车失败!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"updateCart-compensate": "SUCCESS", "Message": "回滚购物车成功!"})
	})

	router.POST("/addTrade", func(c *gin.Context) {
		req := &proto.AddTradeOrderReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		_, err := AddTraderClient.AddTradeOrder(context.TODO(), req)
		if err != nil {
			log.Println("/addTrade err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "新增订单失败!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"addTrade": "SUCCESS", "Message": "新增订单成功!"})
	})
	router.POST("/addTrade-compensate", func(c *gin.Context) {
		req := &proto.AddTradeOrderReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		req.TradeOrder.IsDeleted = true
		_, err := UpdateTraderClient.UpdateTradeOrder(context.TODO(), req)
		if err != nil {
			log.Println("/addTrade err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "回滚订单失败!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"addTrade-compensate": "SUCCESS", "Message": "回滚订单成功!"})
	})

	//新增订单API
	router.GET("/cartAdvanceOrder", func(c *gin.Context) {
		//开始检验登录
		uuid := c.GetHeader("Uuid")
		if uuid == "" {
			common.RespFail(c.Writer, resp, "请在请求头 Uuid 里带上登录返回的用户 ID，例如 2")
			return
		}
		cc := common.GetInput(uuid)
		out := common.SQ(cc)
		sum := 0
		for o := range out {
			sum += o
		}
		//Token校验
		//拼接请求信息
		tokenReq := &proto.TokenReq{
			Uuid: uuid,
		}
		//响应
		tokenResp, err := GetUserTokenClient.GetUserToken(context.TODO(), tokenReq)
		if err != nil || tokenResp.IsLogin == false {
			log.Println("GetUserToken  err : ", err)
			common.RespFail(c.Writer, tokenResp, "未登录！")
			return
		}
		//拼接请求信息
		log.Println("GetUserToken success : ", tokenResp)
		//结束检验登录

		//开始订单插入
		tempStr := c.Request.FormValue("cartIds") // 12,355,666
		cartIds := common.SplitToInt32List(tempStr, ",")
		if len(cartIds) == 0 {
			common.RespFail(c.Writer, resp, "cartIds 不能为空，例如 cartIds=1")
			return
		}
		isVirtual, _ := strconv.ParseBool(c.Request.FormValue("isVirtual"))
		recipientAddressId, _ := strconv.Atoi(c.Request.FormValue("recipientAddressId"))

		//统计价格
		totalReq := &proto.OrderTotalReq{
			CartIds: cartIds,
		}

		//开始校验 cart
		findCartReq := &proto.FindCartReq{
			Id: cartIds[0],
		}
		cart, err := FindCartClient.FindCart(context.TODO(), findCartReq)
		if err != nil {
			log.Println("FindCart  err : ", err)
			common.RespFail(c.Writer, tokenResp, "查询购物车失败！")
			return
		}
		if cart.ShoppingCart.IsDeleted {
			common.RespFail(c.Writer, tokenResp, " 购物车已失效！")
			return
		}
		//结束cart的状态校验
		totalPriceResp, _ := GetOrderTotalClient.GetOrderTotal(context.TODO(), totalReq)
		log.Println("totalPrice>>>>>>>>>>>>>>>>>>>>>>>>   ", totalPriceResp)
		//构建tradeOrder
		tradeOrder := &proto.TradeOrder{}
		tradeOrder.UserId = int32(sum)
		tradeOrder.CreateUser = int32(sum)
		tradeOrder.OrderStatus = 1
		tradeOrder.TotalAmount = totalPriceResp.TotalPrice
		req := &proto.AddTradeOrderReq{
			CartIds:            cartIds,
			IsVirtual:          isVirtual,
			RecipientAddressId: int32(recipientAddressId),
			TradeOrder:         tradeOrder,
		}
		updateCartReq := &proto.UpdateCartReq{
			Id: cartIds[0],
		}

		//全局事务
		gid := shortuuid.New()
		saga := dtmcli.NewSaga(common.DtmServer, gid).
			Add(common.QSBusi+"/updateCart", common.QSBusi+"/updateCart-compensate", updateCartReq).
			Add(common.QSBusi+"/addTrade", common.QSBusi+"/addTrade-compensate", req)
		err = saga.Submit()
		if err != nil {
			log.Println("saga submit err :", err)
			common.RespFail(c.Writer, resp, "添加失败，请确认 DTM 已在 36789 端口启动")
			return
		}
		log.Println(" /saga submit   submit  :", gid)
		////writer  data  message  row  total  field
		common.RespOK(c.Writer, resp, "请求成功")

	})
	router.POST("/findOrder", func(c *gin.Context) {
		req := &proto.FindOrderReq{}
		req.Id = c.PostForm("id")
		req.OrderNo = c.PostForm("orderNo")

		obj, err := FindOrderClient.FindOrder(context.TODO(), req)
		if err != nil {
			log.Println("findOrder err :", err)
			common.RespFail(c.Writer, resp, "查询失败")
		}
		fmt.Println("findOrder:", obj)
		c.JSON(http.StatusOK, gin.H{"findOrder": "SUCCESS", "Message": obj})
	})

	service := web.NewService(webOpts...)
	//启动服务
	service.Run()
}

type clientWrapper struct {
	client.Client
}

func (c clientWrapper) Call(ctx context.Context, req client.Request, resp interface{}, opts ...client.CallOption) error {
	return hystrix.Do(req.Service()+"."+req.Endpoint(), func() error {
		//正常执行
		fmt.Println("call success ", req.Service()+"."+req.Endpoint())
		return c.Client.Call(ctx, req, resp, opts...)
	}, func(err error) error {
		fmt.Println("call err :", err)
		return err
	})

}

func NewClientHystrixWrapper() client.Wrapper {
	return func(i client.Client) client.Client {
		return &clientWrapper{i}
	}

}
