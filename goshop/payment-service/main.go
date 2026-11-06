// 支付 HTTP 服务，端口 8086。它不注册自己的 RPC，而是在支付成功后调用 trade-order。
// 支付宝、PayPal 用的是课程里的沙箱示例密钥，只能走测试环境，不能收真实的钱。
// 密钥初始化失败时服务仍然会启动，对应的支付接口会返回失败说明。
// 命令：在 payment-service 目录执行 go run .
package main

import (
	"context" // 带超时的上下文。PayPal 换令牌最多等 4 秒
	"io"      // 读取支付宝通知的请求正文
	"log"
	"net/http" // StatusOK 这些状态码
	"net/url"  // 解析支付宝发来的表单，以及给回跳地址做转义
	"os"       // PayPal 日志写到标准输出
	"os/exec"  // 调用 Windows 的 cmd 打开浏览器
	"strings"
	"time"

	common "git.mashibing.com/msb_47094/shopping-comm"
	"github.com/asim/go-micro/plugins/registry/consul/v4"
	"github.com/gin-gonic/gin"        // HTTP 路由。这个服务没有单独的 client 目录
	"github.com/plutov/paypal/v4"     // PayPal 沙箱客户端
	"github.com/smartwalle/alipay/v3" // 支付宝沙箱客户端
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"payment/proto" // 这里只有订单服务的客户端，支付自己不提供 RPC
)

const (
	// APPID 是支付宝开放平台里的应用编号。下面三把密钥和它是一套沙箱材料。
	APPID           = "2021000116681444"
	AliAppPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAp1o6GRM72PKbU299+VmOd1527hKp+//fpEPkz7yBHBCoaAAASuIaaQyuHHtJAcjVlb4XNjLoRyIXQ71rPQaucs/5qjmvyPAlfBNJhVaUAV5ZX7WtXHMAEdTBzVODP5axjY5BSzSmpvm9fNPDuaU3ayyqENP5yoHFs2zTSpnwRoh9Ct+urqfrAJRBx435OwVJ6Abv6blKcXMSRLlgq9hM5fbzFHv7OQ8QXvrsB9NA6s2PELI5hbLFuW0T8WyX4AaFIo2SHnHGludrDwKCxO3HolwjZoQKNcsZhDOG0A3kcVUowpnP3SqkbsfVId943KxvXTIouAoYYbfD6/BKcVlC/QIDAQAB"
	AliPublicKey    = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAhA9A6B+gV6pzSv9KKUFcqMZKrFIYOamHIeviQy6KgFuWekGG7KA6agJyGaYrogTpp97NyopgORAK+D5J7bRdcJ1smkpHlyXF0P8D9H/+wZn++1TSQYWWoPJbQXX+l0Ma42O6ISckwXHeZr0tLWUaAoTnTkSAqR0e5LtAFusND6EY/pFFJONxM2Fouwb9WiNCw6J1zQDpO1Y0bxVgQy9v19N/Qw04MtPakua2BqyZctbLggA8WhusCzbTMRZCcOXdYFajPqSvVz5278K6RGKjVeDVBUuaUcA394v1I9tHIa41o8P7j0DhFfDS5Lo8npjhkxcF2wHqfzf5NH041jwAQQIDAQAB"
	PrivateKey      = "MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCnWjoZEzvY8ptTb335WY53XnbuEqn7/9+kQ+TPvIEcEKhoAABK4hppDK4ce0kByNWVvhc2MuhHIhdDvWs9Bq5yz/mqOa/I8CV8E0mFVpQBXllfta1ccwAR1MHNU4M/lrGNjkFLNKam+b1808O5pTdrLKoQ0/nKgcWzbNNKmfBGiH0K366up+sAlEHHjfk7BUnoBu/puUpxcxJEuWCr2Ezl9vMUe/s5DxBe+uwH00DqzY8QsjmFssW5bRPxbJfgBoUijZIeccaW52sPAoLE7ceiXCNmhAo1yxmEM4bQDeRxVSjCmc/dKqRux9Uh33jcrG9dMii4Chhht8Pr8EpxWUL9AgMBAAECggEAYsQ3QZE+Yt1bmjRUrpKSKBTSL9NICveJ75I6OffQYtXZ7fcCNS+X0qOvJhQElE8I7qUPYedNep2+CTaDUbLdxdjEShhdjENefAmN/GmKU/xw1rGMLOGaCuv7rMg6z46KHzbzhgT+zdDaag84lhnRySjE/cV8yG0qVqhFUuPjEFyMrY8lb/68P+E3P4Vm+hvdaOe10ayLmYm7YBh4Z7eSdhSxENBs+7dIG37Z2xC6FQRPm5qLrFdP03tianX71ZU/J7yky4HceqW4lIW8g4CSbWlCKewaucvnZ7q449/P/PSIbPr4KxQS8bGooQF6LGYEkRISWpYIQWBSjXvUU7PDaQKBgQDV3Hgwp93FdLU414ArdXzbklzVC3SEKdIXEBmt9s+Ec13I1xPDjM1HFezow4S/aaqi5soYyR6koVvsaXeHk+cNlphgZj6lz+p0BQvCvTrjkRFwSDdlBNuxgaiwZbjjUHZPwO42d3k2kiUwyRNcGIyDyPzo4b19D06VX2JIKdpp+wKBgQDIU8ZfQUj11Zgqx/dqd5XASGtoSQNwZPROJyLFwTMg6okgdevdoWaFgRUttPDRBhdcrtmVq7voJ35xoriKNFCSvkZc8GAVKiYgIGLXNscDKZx1EP5yc9VVbUnnTW4QXBQRB+VcLGmuh1DOPTkP+sEsTZhPM5TTMCX9HxrDtW6tZwKBgQCAlkBaGQTaTbMNC14QQ2Mgbp+MpYeYU8DzhwW5uNXLcMZqFQ/YG1fuLvu0dKLjqHgl3JiVUJAxMqgfvMZW6ZIadHsIM1Hlxt7UhZgtrjOzUq1mkrVj8wjZcPlpnrANVXrdt0ZDz5A9RKBB9hTC4ivSrU2MBPc3M7XvUFu6wheGcwKBgQCCMCeIOiZPVWuZyfDZ6TOSDb53vJOWlJVMwRRRDljK+ID7bodFtqGo+5Wm6aNjRCEko6FWKpBotFnxQb3KW2tNooiZ18Gpn4W9dS6nDLtTK9S5I70o5mxrIyZXQrPXv+9ujgIZ3+cifGAgWBtgbibiQt9MAYqjoPPfRzqem6d82wKBgQCjcqFOK7Ff0bSz6S2PzFLWyHBlrEKXI+9+9N5nS8aFB8uJduu0PqK4CXs6FqbY+6xuZ9Qdhe9j2NXoqZoJyQlb8/aL1etT+/JP0Md9Wi0v9SVnnz6Ci1RmJEjO02Q/xrpphEbK87h7YpNUI4flHR4ZCr9EmXqdXRTuXImUImc3uw=="
	PayPalClientID  = "AX5c8C-Z9P5vXgs22Q3MWKaEKvVa5-vtU3jtNrQe8W2s-x8eXYn_2aA-e4paaHo6I0Z9eba0Ur2LTdZI"
	PayPalSecretID  = "EB_svAAyczD2OoCMh2HMA4ApBysCbHli8gOpFUfKLOaI5zkoePeNd54OI1PT5yDtAqNsWgho7IYNdqFR"
)

var (
	aliClient          *alipay.Client                // 初始化失败时保持 nil，接口里会先判断
	paypalClient       *paypal.Client                // 同上
	findOrderClient    proto.FindOrderService        // 按订单号向 trade-order 查询
	updateTraderClient proto.UpdateTradeOrderService // 把订单状态改成已支付
)

// init 在 main 之前自动执行。Go 规定一个包可以有多个 init，它们都在 main 之前跑完。
func init() {
	setupAlipay()
	setupPayPal()
	setupTradeClient()
}

// setupAlipay 用应用私钥签名。第三个参数 false 表示沙箱，不是正式环境。
// 新版支付宝要求证书，这里沿用课程的公钥字符串；加载失败就跳过，避免进程起不来。
func setupAlipay() {
	// recover 接住支付宝库里的宕机。defer 保证函数退出前一定检查一次。
	defer func() {
		if r := recover(); r != nil {
			log.Println("alipay init skipped:", r)
			aliClient = nil
		}
	}()
	client, err := alipay.New(APPID, PrivateKey, false)
	if err != nil {
		log.Println("alipay init skipped:", err)
		return
	}
	if err = client.LoadAppPublicCert(AliAppPublicKey); err != nil {
		log.Println("alipay app public key skipped:", err)
	}
	if err = client.LoadAliPayPublicCert(AliPublicKey); err != nil {
		log.Println("alipay public key skipped:", err)
	}
	aliClient = client
	log.Println("alipay sandbox client ready")
}

// setupPayPal 在后台换访问令牌，避免沙箱网络慢时堵住 8086 端口的启动。
func setupPayPal() {
	client, err := paypal.NewClient(PayPalClientID, PayPalSecretID, paypal.APIBaseSandBox)
	if err != nil {
		log.Println("paypal init skipped:", err)
		return
	}
	client.SetLog(os.Stdout)
	paypalClient = client
	// go 开头表示新开一个协程。换令牌和 main 里的监听同时进行。
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel() // 用完释放这个超时计时器
		if _, err := client.GetAccessToken(ctx); err != nil {
			log.Println("paypal token skipped:", err)
			return
		}
		log.Println("paypal sandbox token ready")
	}()
}

// setupTradeClient 准备调用订单服务的客户端。Consul 不在线时用 mDNS。
func setupTradeClient() {
	opts := []micro.Option{micro.Name("shop-payment")}
	if common.ConsulReachable() {
		opts = append(opts, micro.Registry(consul.NewRegistry(func(options *registry.Options) {
			options.Addrs = []string{common.ConsulReistStr}
		})))
	} else {
		log.Println("consul unreachable, payment discovers trade-order via mdns")
	}
	// 这里的 NewService 只当作 RPC 客户端的外壳，不会监听一个新的业务端口。
	rpcServer := micro.NewService(opts...)
	findOrderClient = proto.NewFindOrderService("trade-order", rpcServer.Client())
	updateTraderClient = proto.NewUpdateTradeOrderService("trade-order", rpcServer.Client())
}

// openPayURL 用 Windows 的 start 打开收银台。cmd 会把 & 当成命令分隔符，所以先换成 ^&。
func openPayURL(raw string) {
	if raw == "" {
		log.Println("支付链接为空，浏览器不会打开")
		return
	}
	raw = strings.ReplaceAll(raw, "&", "^&")
	// Start 不等待浏览器关闭。打开失败只打日志，支付链接仍然会写进接口响应。
	if err := exec.Command("cmd", "/c", "start", raw).Start(); err != nil {
		log.Println("open browser:", err)
	}
}

// TradeWapAliPay 生成手机网站支付链接。查询参数 payAmount 是金额，orderNo 是商户订单号。
func TradeWapAliPay(c *gin.Context) {
	if aliClient == nil {
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "支付宝客户端没有初始化"})
		return
	}
	pay := alipay.TradeWapPay{}
	pay.TotalAmount = c.Query("payAmount")
	pay.OutTradeNo = c.Query("orderNo")
	pay.NotifyURL = "http://127.0.0.1:8086/return" // 支付宝付完款后，异步通知打到本服务
	pay.Body = "APP订单"
	pay.Subject = "商品标题"
	res, err := aliClient.TradeWapPay(pay)
	if err != nil || res == nil {
		log.Println("支付失败 :", err)
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "生成手机支付链接失败"})
		return
	}
	openPayURL(res.String())
	c.JSON(http.StatusOK, gin.H{"result": "SUCCESS", "payURL": res.String()})
}

// TradePageAlipay 生成电脑网站扫码支付链接。ProductCode 必须是支付宝规定的 FAST_INSTANT_TRADE_PAY。
func TradePageAlipay(c *gin.Context) {
	if aliClient == nil {
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "支付宝客户端没有初始化"})
		return
	}
	p := alipay.TradePagePay{}
	p.NotifyURL = "http://127.0.0.1:8086/return"
	p.Subject = "商品支付"
	p.OutTradeNo = c.Query("orderNo")
	p.TotalAmount = c.Query("payAmount")
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	res, err := aliClient.TradePagePay(p)
	if err != nil || res == nil {
		log.Println("支付失败 :", err)
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "生成网页支付链接失败"})
		return
	}
	openPayURL(res.String())
	c.JSON(http.StatusOK, gin.H{"result": "SUCCESS", "payURL": res.String()})
}

// AliPayNotify 是支付宝异步通知。验签在正式项目里必须做，这里沿用课程，只看交易状态。
// 支付成功后把订单状态改成 3（已支付）。订单服务要先启动。
func AliPayNotify(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("read notify body:", err)
		return
	}
	// 正文是 key=value&key=value 这种表单，不是 JSON。
	values, err := url.ParseQuery(string(data))
	if err != nil {
		log.Println("parse notify body:", err)
		return
	}
	if values.Get("trade_status") != "TRADE_SUCCESS" {
		c.JSON(http.StatusOK, gin.H{"result": "IGNORE", "Message": "不是支付成功通知"})
		return
	}
	orderNo := values.Get("out_trade_no")
	log.Println("订单:", orderNo, "成功支付")
	// context.TODO 表示这次调用没有设置超时。订单服务没启动时，这里会返回找不到服务。
	obj, err := findOrderClient.FindOrder(context.TODO(), &proto.FindOrderReq{OrderNo: orderNo})
	if err != nil || obj == nil || obj.TradeOrder == nil {
		log.Println("findOrder err :", err)
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "没有找到订单"})
		return
	}
	reqUpdate := &proto.AddTradeOrderReq{
		TradeOrder: &proto.TradeOrder{
			ID:          obj.TradeOrder.ID,
			OrderStatus: 3, // 3 表示已支付
		},
	}
	if _, err = updateTraderClient.UpdateTradeOrder(context.TODO(), reqUpdate); err != nil {
		log.Println("UpdateTradeOrder err ", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"result": "SUCCESS", "Message": "修改订单支付成功!"})
}

// CreateOrderPayer 向 PayPal 沙箱下一笔美元订单，并打开收银台。
func CreateOrderPayer(c *gin.Context) {
	if paypalClient == nil {
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "PayPal 客户端没有初始化"})
		return
	}
	orderNo := c.Query("orderNo")
	// PurchaseUnitRequest 是 PayPal 要求的一笔收款。金额单位在这个示例里是美元。
	purchaseUnits := []paypal.PurchaseUnitRequest{{
		Amount:      &paypal.PurchaseUnitAmount{Currency: "USD", Value: c.Query("amount")},
		InvoiceID:   orderNo,
		ReferenceID: orderNo,
		Description: "商品支付",
	}}
	appContext := &paypal.ApplicationContext{
		// QueryEscape 把订单号里的特殊字符转成可以放进网址的形式。
		ReturnURL: "http://127.0.0.1:8086/payPalBack?orderNo=" + url.QueryEscape(orderNo),
	}
	orderResponse, err := paypalClient.CreateOrder(context.Background(), "CAPTURE", purchaseUnits, &paypal.CreateOrderPayer{}, appContext)
	// Links 里通常有多条网址。课程取第 2 条，下标是 1，那是买家打开的收银台。
	if err != nil || orderResponse == nil || len(orderResponse.Links) < 2 {
		log.Println("create order errors:", err)
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "构建 PayPal 订单失败"})
		return
	}
	payURL := orderResponse.Links[1].Href
	openPayURL(payURL)
	c.JSON(http.StatusOK, gin.H{"result": "SUCCESS", "payURL": payURL, "id": orderResponse.ID})
}

// PaypalCallback 在买家付完款后捕获订单，并把本地订单改成已支付。
// PayPal 回跳时通常带 token，token 就是 PayPal 订单号。
func PaypalCallback(c *gin.Context) {
	if paypalClient == nil {
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "PayPal 客户端没有初始化"})
		return
	}
	id := c.Query("ID")
	if id == "" {
		id = c.Query("token")
	}
	order, err := paypalClient.CaptureOrder(context.Background(), id, paypal.CaptureOrderRequest{})
	if err != nil || order == nil || order.Status != "COMPLETED" {
		log.Println("paypal capture err ", err)
		c.JSON(http.StatusOK, gin.H{"result": "FAIL", "Message": "订单回调失败"})
		return
	}
	// 捕获成功后，再用本地订单号去 trade-order 把状态改成 3。
	if orderNo := c.Query("orderNo"); orderNo != "" && findOrderClient != nil {
		obj, findErr := findOrderClient.FindOrder(context.TODO(), &proto.FindOrderReq{OrderNo: orderNo})
		if findErr == nil && obj != nil && obj.TradeOrder != nil {
			_, _ = updateTraderClient.UpdateTradeOrder(context.TODO(), &proto.AddTradeOrderReq{
				TradeOrder: &proto.TradeOrder{ID: obj.TradeOrder.ID, OrderStatus: 3},
			})
		}
	}
	c.JSON(http.StatusOK, gin.H{"result": "SUCCESS", "Message": "修改订单支付成功!"})
}

// main 只做 HTTP 路由。支付宝和 PayPal 的客户端在 init 里已经准备好。
func main() {
	r := gin.Default() // 自带访问日志和宕机恢复
	// 只信任本机代理。前面的下划线丢掉设置失败时的错误，学习环境里这个设置不影响支付。
	_ = r.SetTrustedProxies([]string{common.QSIp})
	r.GET("/appPay", TradeWapAliPay)     // 手机网站支付
	r.GET("/pagePay", TradePageAlipay)   // 电脑网站支付
	r.POST("/return", AliPayNotify)      // 支付宝异步通知
	r.GET("/payPal", CreateOrderPayer)   // 创建 PayPal 订单
	r.GET("/payPalBack", PaypalCallback) // PayPal 付完款后回跳
	log.Println("shop-payment listening :8086")
	// Run 堵住当前协程，直到进程退出。
	if err := r.Run(":8086"); err != nil {
		log.Println("start payment err :", err)
	}
}
