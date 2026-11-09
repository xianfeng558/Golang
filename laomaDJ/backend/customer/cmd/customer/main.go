package main

import (
	"flag"
	"os"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSDK "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"customer/internal/conf"
	"customer/internal/discover"

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
	// Name 是注册到 Consul 时的服务名，也是别的进程用 discovery:///Customer 查找时的名字。
	// 改名后，乘客自己能起来，但已经写死 Customer 的调用方会找不到它。
	Name string = "Customer"
	// Version 只写在注册信息和日志里，不参与接口兼容判断。
	Version string = "1.0.0"
	// flagconf 是配置目录。默认 ../../configs，只在当前目录是 cmd/customer 时才指到对的地方。
	// 在模块根目录启动时要写成 -conf ./configs。
	flagconf string

	//id, _ = os.Hostname()
	id = Name + "-" + uuid.NewString()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	helper := log.NewHelper(logger)
	// Jaeger 没开时导出会失败，不影响接口。
	if err := initTracer("http://127.0.0.1:14268/api/traces"); err != nil {
		helper.Errorf("jaeger tracer skipped: %v", err)
	}

	opts := []kratos.Option{
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	}
	reg, err := discover.Registrar()
	if err != nil {
		helper.Errorf("consul registrar: %v", err)
	} else if reg != nil {
		opts = append(opts, kratos.Registrar(reg))
	} else {
		helper.Info("consul unreachable, skip register; rpc uses direct address")
	}
	return kratos.New(opts...)
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

// initTracer 把请求追踪发到本机 Jaeger。
// url 是采集器地址，学习环境写死 http://127.0.0.1:14268/api/traces。
// Jaeger 没启动时这里返回错误，调用方只记日志，登录和询价照常工作。
// AlwaysSample 表示每条请求都采样，只适合本机。
func initTracer(url string) error {
	// exporter 把 span 批量发给 Jaeger。
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return err
	}
	// TracerProvider 决定采样比例、批量发送，以及这条链路属于哪个服务名。
	tracerProvider := traceSDK.NewTracerProvider(
		traceSDK.WithSampler(traceSDK.AlwaysSample()),
		traceSDK.WithBatcher(exporter),
		traceSDK.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
			attribute.String("exporter", "jaeger"),
		)),
	)

	// 设成全局之后，HTTP 上的 tracing 中间件才会把 span 写到这个导出器。
	otel.SetTracerProvider(tracerProvider)

	return nil
}
