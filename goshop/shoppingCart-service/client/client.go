// 购物车 HTTP 网关，端口 6668。
// 它自己不查数据库，而是用 go-micro 客户端去调用 shop-cart、shop-product、shop-user。
// 启动前要先把这三个 RPC 服务跑起来，否则这里的接口会返回失败。
package main

import (
	"context"
	common "git.mashibing.com/msb_47094/shopping-comm"
	"github.com/afex/hystrix-go/hystrix"
	consul "github.com/asim/go-micro/plugins/registry/consul/v4"
	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/go-micro/plugins/v4/wrapper/select/roundrobin"
	opentracing2 "github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	"github.com/lithammer/shortuuid/v3"
	"github.com/opentracing/opentracing-go"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/web"
	"log"
	"net"
	"net/http"
	"shoppingCart/proto"
	"strconv"

	"github.com/gin-gonic/gin"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
)

// 获取远程服务的客户端
func main() {
	var CartId int32 = 1
	var Number int32 = 1
	resp := &proto.AddCartResp{}

	router := gin.Default()

	// 没有 Consul 时不要注册，否则客户端只会去一个不存在的注册中心找服务。
	var consulReg registry.Registry
	if common.ConsulReachable() {
		consulReg = consul.NewRegistry(func(options *registry.Options) {
			options.Addrs = []string{common.ConsulReistStr}
		})
	} else {
		log.Println("consul unreachable, discover services via mdns")
	}
	if closer := common.UseTracer("shop-cart-client"); closer != nil {
		defer closer.Close()
	}

	// 熔断器把失败的下游调用快速挡掉。9096 是它的监控页，打不开不影响业务。
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	go func() {
		err := http.ListenAndServe(net.JoinHostPort("127.0.0.1", "9096"), hystrixStreamHandler)
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
		web.Address(":6668"),
		web.Name("shop-cart-client"),
		web.Handler(router),
	}
	if consulReg != nil {
		opts = append(opts, micro.Registry(consulReg))
		webOpts = append(webOpts, web.Registry(consulReg))
	}
	rpcServer := micro.NewService(opts...)
	AddCartClient := proto.NewAddCartService("shop-cart", rpcServer.Client())
	UpdateCartClient := proto.NewUpdateCartService("shop-cart", rpcServer.Client())
	ShowProductDetailClient := proto.NewShowProductDetailService("shop-product", rpcServer.Client())
	ShowDetailSkuClient := proto.NewShowDetailSkuService("shop-product", rpcServer.Client())
	GetUserTokenClient := proto.NewGetUserTokenService("shop-user", rpcServer.Client())
	UpdateSkuClient := proto.NewUpdateSkuService("shop-product", rpcServer.Client())

	//添加购物车
	router.GET("/increase", func(c *gin.Context) {
		number, _ := strconv.Atoi(c.Request.FormValue("number"))
		productId, _ := strconv.Atoi(c.Request.FormValue("productId"))
		productSkuId, _ := strconv.Atoi(c.Request.FormValue("productSkuId"))
		uuid := c.GetHeader("Uuid")
		if uuid == "" {
			common.RespFail(c.Writer, &proto.AddCartResp{}, "请在请求头 Uuid 里带上登录返回的用户 ID，例如 2")
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
		//拼接请求信息
		respErr := &proto.AddCartResp{}
		if err != nil || tokenResp.IsLogin == false {
			log.Println("GetUserToken  err : ", err)
			common.RespFail(c.Writer, respErr, "未登录！")
			return
		}
		log.Println("GetUserToken success : ", tokenResp)

		//拼接请求信息
		req := &proto.AddCartReq{
			Number:       int32(number),
			ProductId:    int32(productId),
			ProductSkuId: int32(productSkuId),
			UserId:       int32(sum),
		}
		resp := &proto.AddCartResp{}
		//商品详情
		reqDetail := &proto.ProductDetailReq{
			Id: int32(productId),
		}
		respDetail, err := ShowProductDetailClient.ShowProductDetail(context.TODO(), reqDetail)
		if err != nil {
			log.Println("ShowProductDetail  err : ", err)
			common.RespFail(c.Writer, respErr, "查询商品详情失败！")
			return
		}
		if respDetail == nil || len(respDetail.ProductDetail) == 0 {
			common.RespFail(c.Writer, respErr, "商品不存在")
			return
		}
		req.ProductName = respDetail.ProductDetail[0].Name
		req.ProductMainPicture = respDetail.ProductDetail[0].MainPicture

		//log.Println(" /ShowProductDetail   resp   :", respDetail)
		//SKU详情
		reqDetail.Id = req.ProductSkuId
		respSkuDetail, err := ShowDetailSkuClient.ShowDetailSku(context.TODO(), reqDetail)
		//log.Println(" /ShowDetailSku   resp   :", respSkuDetail)
		//添加购物车  远程调用服务
		//log.Println(" /AddCart  req :", req)

		if err != nil || respSkuDetail == nil || len(respSkuDetail.ProductSku) == 0 {
			common.RespFail(c.Writer, respErr, "查询规格失败")
			return
		}
		if respSkuDetail.ProductSku[0].Stock < req.Number {
			common.RespFail(c.Writer, &proto.AddCartResp{}, "库存不足，添加失败")
			return
		}
		sku := respSkuDetail.ProductSku[0]
		sku.Stock -= req.Number
		updateSkuReq := &proto.UpdateSkuReq{
			ProductSku: sku,
		}
		respUpdate, err := UpdateSkuClient.UpdateSku(context.TODO(), updateSkuReq)
		if err != nil {
			log.Println(" /UpdateSku  err :", err)
			common.RespFail(c.Writer, resp, "修改库存失败！")
			return
		}
		log.Println(" /UpdateSkuClient  resp :", respUpdate.IsSuccess)
		//开始增加购物车
		resp, err = AddCartClient.AddCart(context.TODO(), req)
		//根据响应做输出
		if err != nil {
			log.Println("addCart err ", err)
			updateSkuReq.ProductSku.Stock += req.Number
			_, err = UpdateSkuClient.UpdateSku(context.TODO(), updateSkuReq)
			log.Println("rollback sku  is Err :", err)
			common.RespFail(c.Writer, resp, "添加购物车失败！")
			return
		}
		resp.ProductSkuSimple = respSkuDetail.ProductSku[0]
		resp.ProductSimple = respDetail.ProductDetail[0]
		log.Println(" /AddCart  resp :", resp)
		////writer  data  message  row  total  field
		common.RespOK(c.Writer, resp, "请求成功")
	})

	//开始拆分 DTM服务
	//
	router.POST("/updateSku", func(c *gin.Context) {
		req := &proto.UpdateSkuReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		_, err := UpdateSkuClient.UpdateSku(context.TODO(), req)
		if err != nil {
			log.Println("/updateSku err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "修改库存失败!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"updateSku": "SUCCESS", "Message": "修改库存成功!"})
	})
	router.POST("/updateSku-compensate", func(c *gin.Context) {
		req := &proto.UpdateSkuReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		req.ProductSku.Stock += Number
		_, err := UpdateSkuClient.UpdateSku(context.TODO(), req)
		if err != nil {
			log.Println("/updateSku err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "回滚库存失败!"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"updateSku-compensate": "SUCCESS", "Message": "回滚库存成功!"})
	})
	router.POST("/addCart", func(c *gin.Context) {
		req := &proto.AddCartReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		added, err := AddCartClient.AddCart(context.TODO(), req)
		if err != nil || added == nil {
			log.Println("/addCart err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "新增购物车失败!"})
			return
		}
		resp = added
		CartId = resp.ID

		c.JSON(http.StatusOK, gin.H{"addCart": "SUCCESS", "Message": "新增购物车成功!"})
	})
	router.POST("/addCart-compensate", func(c *gin.Context) {
		req := &proto.UpdateCartReq{}
		if err := c.BindJSON(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"dtm_reslut": "FAILURE", "Message": err.Error()})
			return
		}
		req.Id = CartId
		updated, err := UpdateCartClient.UpdateCart(context.TODO(), req)
		if err != nil || updated == nil {
			log.Println("/addCart-compensate err ", err)
			c.JSON(http.StatusOK, gin.H{"dtm_reslut": "FAILURE", "Message": "删除购物车失败!"})
			return
		}
		CartId = updated.ID
		c.JSON(http.StatusOK, gin.H{"addCart-compensate": "SUCCESS", "Message": "删除购物车成功!"})
	})

	router.GET("/addShoppingCart", func(c *gin.Context) {
		number, _ := strconv.Atoi(c.Request.FormValue("number"))
		productId, _ := strconv.Atoi(c.Request.FormValue("productId"))
		productSkuId, _ := strconv.Atoi(c.Request.FormValue("productSkuId"))
		uuid := c.GetHeader("Uuid")
		if uuid == "" {
			common.RespFail(c.Writer, &proto.AddCartResp{}, "请在请求头 Uuid 里带上登录返回的用户 ID，例如 2")
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
		//拼接请求信息
		respErr := &proto.AddCartResp{}
		if err != nil || tokenResp.IsLogin == false {
			log.Println("GetUserToken  err : ", err)
			common.RespFail(c.Writer, respErr, "未登录！")
			return
		}
		log.Println("GetUserToken success : ", tokenResp)

		//拼接请求信息
		req := &proto.AddCartReq{
			Number:       int32(number),
			ProductId:    int32(productId),
			ProductSkuId: int32(productSkuId),
			UserId:       int32(sum),
			CreateUser:   int32(sum),
		}
		resp := &proto.AddCartResp{}
		//商品详情
		reqDetail := &proto.ProductDetailReq{
			Id: int32(productId),
		}
		respDetail, err := ShowProductDetailClient.ShowProductDetail(context.TODO(), reqDetail)
		if err != nil {
			log.Println("ShowProductDetail  err : ", err)
			common.RespFail(c.Writer, respErr, "查询商品详情失败！")
			return
		}
		if respDetail == nil || len(respDetail.ProductDetail) == 0 {
			common.RespFail(c.Writer, respErr, "商品不存在")
			return
		}
		req.ProductName = respDetail.ProductDetail[0].Name
		req.ProductMainPicture = respDetail.ProductDetail[0].MainPicture

		//log.Println(" /ShowProductDetail   resp   :", respDetail)
		//SKU详情
		reqDetail.Id = req.ProductSkuId
		respSkuDetail, err := ShowDetailSkuClient.ShowDetailSku(context.TODO(), reqDetail)
		//log.Println(" /ShowDetailSku   resp   :", respSkuDetail)
		//添加购物车  远程调用服务
		//log.Println(" /AddCart  req :", req)

		if err != nil || respSkuDetail == nil || len(respSkuDetail.ProductSku) == 0 {
			common.RespFail(c.Writer, respErr, "查询规格失败")
			return
		}
		if respSkuDetail.ProductSku[0].Stock < req.Number {
			common.RespFail(c.Writer, &proto.AddCartResp{}, "库存不足，添加失败")
			return
		}
		sku := respSkuDetail.ProductSku[0]
		sku.Stock -= req.Number
		Number = req.Number //
		updateSkuReq := &proto.UpdateSkuReq{
			ProductSku: sku,
		}
		resp.ProductSkuSimple = respSkuDetail.ProductSku[0]
		resp.ProductSimple = respDetail.ProductDetail[0]

		//全局事务
		gid := shortuuid.New()
		// DTM 的回调必须打到本网关。购物车接口在 6668，不能误用订单网关的 6669。
		saga := dtmcli.NewSaga(common.DtmServer, gid).
			Add(common.CartHTTP+"/updateSku", common.CartHTTP+"/updateSku-compensate", updateSkuReq).
			Add(common.CartHTTP+"/addCart", common.CartHTTP+"/addCart-compensate", req)
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

	service := web.NewService(webOpts...)
	//启动服务
	service.Run()
}

type clientWrapper struct {
	client.Client
}

// Call 给每次 RPC 套上熔断。下游连续失败时快速返回，不再一直等待。
func (c clientWrapper) Call(ctx context.Context, req client.Request, resp interface{}, opts ...client.CallOption) error {
	return hystrix.Do(req.Service()+"."+req.Endpoint(), func() error {
		return c.Client.Call(ctx, req, resp, opts...)
	}, func(err error) error {
		log.Println("hystrix:", req.Service()+"."+req.Endpoint(), err)
		return err
	})
}

func NewClientHystrixWrapper() client.Wrapper {
	return func(i client.Client) client.Client {
		return &clientWrapper{i}
	}

}
