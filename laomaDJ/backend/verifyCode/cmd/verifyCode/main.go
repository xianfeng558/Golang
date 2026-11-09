package main

import (
	"flag"
	"net"
	"os"
	"time"

	consulreg "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"

	"verifyCode/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name 必须是 VerifyCode。乘客和司机按这个名字要验证码，直连时则固定访问 127.0.0.1:9000。
	Name string = "VerifyCode"
	// Version 只写在注册信息和日志里。
	Version string = "1.0.0"
	// flagconf 默认 ../../configs。在 verifyCode 模块根目录启动时要写成 -conf ./configs。
	flagconf string

	//id, _ = os.Hostname()
	// 使用唯一的uuid，作为id
	id = Name + "-" + uuid.NewString()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	helper := log.NewHelper(logger)
	opts := []kratos.Option{
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{"weight": "999"}),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	}
	reg, err := consulRegistrar()
	if err != nil {
		helper.Errorf("consul registrar: %v", err)
	} else if reg != nil {
		opts = append(opts, kratos.Registrar(reg))
	} else {
		helper.Info("consul unreachable, skip register")
	}
	return kratos.New(opts...)
}

// consulRegistrar 用 400 毫秒探测 127.0.0.1:8500。没开就返回 nil, nil，验证码服务照常监听 9000。
// Metadata 里的 weight 999 只在 Consul 开着、并且有多份实例时参与加权，单机直连用不到。
func consulRegistrar() (*consulreg.Registry, error) {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:8500", 400*time.Millisecond)
	if err != nil {
		return nil, nil
	}
	_ = conn.Close()
	cfg := api.DefaultConfig()
	cfg.Address = "127.0.0.1:8500"
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return consulreg.New(client), nil
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
