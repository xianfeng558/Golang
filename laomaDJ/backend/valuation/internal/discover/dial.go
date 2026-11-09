// Package discover 决定怎么找到另一个进程，并复用已经拨通的 gRPC 连接。
// 计价服务用它访问 Map，直连地址是 127.0.0.1:9200。
// 乘客、司机服务各有一份相同职责的包，因为它们是独立的 Go 模块，不能互相 import。
package discover

import (
	"context"
	"net"
	"sync"
	"time"

	consulreg "github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

// ConsulAddr 是本机注册中心。只探测这个地址，不读取配置文件。
// 课程代码曾经写过局域网 IP，换一台电脑就会连不上，所以固定为回环地址。
const ConsulAddr = "127.0.0.1:8500"

// Reachable 在 400 毫秒内尝试连接 8500。
// 连上就认为 Consul 可用并立刻关掉这条探测连接。超时或拒绝都返回 false，调用方改走直连。
func Reachable() bool {
	conn, err := net.DialTimeout("tcp", ConsulAddr, 400*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Registrar 给 kratos.App 用。
// Consul 没开时返回 nil, nil，调用方不要把 nil 注册器塞进应用。
// 创建客户端失败时返回错误。真正的服务登记发生在进程开始监听之后。
func Registrar() (*consulreg.Registry, error) {
	if !Reachable() {
		return nil, nil
	}
	cfg := api.DefaultConfig()
	cfg.Address = ConsulAddr
	client, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return consulreg.New(client), nil
}

// Pool 按服务名保存 gRPC 连接。同一进程里对同一个服务只拨一次号。
// 地图进程重启后，缓存里仍是旧连接。这时要把计价进程也重启。
type Pool struct {
	mu   sync.Mutex
	conn map[string]*grpc.ClientConn
}

// Default 是进程内共用的连接池。
var Default = &Pool{conn: map[string]*grpc.ClientConn{}}

// Conn 返回到 service 的连接。
// service 必须和对方 main.go 里的 Name 一致。direct 是 Consul 不可用时的 host:port。
// 拨号使用不会取消的 context。单次 RPC 超时 8 秒，另外还受 config.yaml 的 timeout 限制。
func (p *Pool) Conn(service, direct string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c := p.conn[service]; c != nil {
		return c, nil
	}

	opts := []kgrpc.ClientOption{
		kgrpc.WithTimeout(8 * time.Second),
		kgrpc.WithMiddleware(tracing.Client()),
	}
	if Reachable() {
		cfg := api.DefaultConfig()
		cfg.Address = ConsulAddr
		client, err := api.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		opts = append(opts,
			kgrpc.WithEndpoint("discovery:///"+service),
			kgrpc.WithDiscovery(consulreg.New(client)),
		)
	} else {
		opts = append(opts, kgrpc.WithEndpoint(direct))
	}

	conn, err := kgrpc.DialInsecure(context.Background(), opts...)
	if err != nil {
		return nil, err
	}
	p.conn[service] = conn
	return conn, nil
}
