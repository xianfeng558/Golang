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

	"valuation/internal/conf"
	"valuation/internal/discover"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// 编译时可以用 -ldflags "-X main.Version=x.y.z" 改写版本号
var (
	// Name 必须是 Valuation。乘客服务按这个名字发现计价进程，直连时则固定访问 127.0.0.1:9300。
	Name string = "Valuation"
	// Version 只写在注册信息和日志里。
	Version string = "1.0.0"
	// flagconf 默认 ../../configs。在 valuation 模块根目录启动时要写成 -conf ./configs。
	flagconf string

	// 原来用主机名当实例编号：id, _ = os.Hostname()
	id = Name + "-" + uuid.NewString()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	helper := log.NewHelper(logger)
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

	// 启动，并一直等到停止信号
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// initTracer 把询价链路发到本机 Jaeger。采集器没开时返回错误，计价接口照常工作。
// AlwaysSample 表示每条请求都采样，只适合本机。
func initTracer(url string) error {
	exporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		return err
	}

	// 服务名会写进每条 span，Jaeger 界面上按 Valuation 过滤。
	tracerProvider := traceSDK.NewTracerProvider(
		traceSDK.WithSampler(traceSDK.AlwaysSample()),
		traceSDK.WithBatcher(exporter),
		traceSDK.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(Name),
			attribute.String("exporter", "jaeger"),
		)),
	)
	otel.SetTracerProvider(tracerProvider)

	return nil
}
