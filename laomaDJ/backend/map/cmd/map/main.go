package main

import (
	"flag"
	"net"
	"os"
	"time"

	consulreg "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"

	"map/internal/conf"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name 必须是 Map。计价服务按这个名字发现地图进程，直连时则固定访问 127.0.0.1:9200。
	Name string = "Map"
	// Version 只写在注册信息和日志里。
	Version string = "1.0.0"
	// flagconf 默认 ../../configs。在 map 模块根目录启动时要写成 -conf ./configs。
	flagconf string

	//id, _ = os.Hostname()
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
		kratos.Metadata(map[string]string{}),
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

// consulRegistrar 用 400 毫秒探测 127.0.0.1:8500。
// 端口没有进程时返回 nil, nil。调用方看到 nil 就跳过注册，地图服务仍然监听 9200，计价可以直连。
// 探测成功但创建客户端失败时返回错误，启动日志会记下来，进程继续跑。
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

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
