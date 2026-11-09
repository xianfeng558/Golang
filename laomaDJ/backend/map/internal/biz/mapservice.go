// Package biz 调用高德驾车规划，把第一条路线的距离和时长交给计价服务。
// 本包不连 MySQL，也不注册到 Consul。调用方是计价服务的 gRPC 客户端。
package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// amapKey 是课程里的高德 Web 服务 key，只用于学习驾车距离。
// 它和前端 main.ts 里的 JS key 不是同一把。失效或配额用完时，询价会在这一层失败。
const amapKey = "4b37316ad0582dc767eaefa58cfd3f34"

// MapServiceBiz 调用高德驾车接口。返回值单位是米和秒，都是十进制字符串，和计价服务约定一致。
type MapServiceBiz struct {
	log    *log.Helper
	client *http.Client
}

// NewMapServiceBiz 供 wire 组装。
// HTTP 客户端超时 8 秒。高德没有响应时，请求在这里结束，不会把地图进程一直挂住。
func NewMapServiceBiz(logger log.Logger) *MapServiceBiz {
	return &MapServiceBiz{
		log:    log.NewHelper(logger),
		client: &http.Client{Timeout: 8 * time.Second},
	}
}

// GetDrivingInfo 请求高德驾车规划，只取第一条路线。
// origin、destination 的格式是「经度,纬度」，例如 116.304388,39.984405。顺序是经度在前。
// extensions=base 只要距离和时长，不要逐步导航。响应体最多读 1MB，避免异常大包占满内存。
// status 不是 "1"、没有路线、或距离时长为空，都返回错误，调用方不会把空串当成 0 公里。
func (msbiz *MapServiceBiz) GetDrivingInfo(ctx context.Context, origin, destination string) (string, string, error) {
	if origin == "" || destination == "" {
		return "", "", fmt.Errorf("起点或终点为空")
	}
	query := url.Values{}
	query.Set("origin", origin)
	query.Set("destination", destination)
	query.Set("extensions", "base")
	query.Set("output", "json")
	query.Set("key", amapKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://restapi.amap.com/v3/direction/driving?"+query.Encode(), nil)
	if err != nil {
		return "", "", err
	}
	resp, err := msbiz.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("高德返回 HTTP %d", resp.StatusCode)
	}

	var parsed directionDrivingResp
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", err
	}
	if parsed.Status != "1" {
		info := parsed.Info
		if info == "" {
			info = "高德驾车规划失败"
		}
		return "", "", fmt.Errorf("%s", info)
	}
	if len(parsed.Route.Paths) == 0 {
		return "", "", fmt.Errorf("高德没有返回路线")
	}
	path := parsed.Route.Paths[0]
	if path.Distance == "" || path.Duration == "" {
		return "", "", fmt.Errorf("高德路线缺少距离或时长")
	}
	return path.Distance, path.Duration, nil
}

// directionDrivingResp 只保留计价需要的字段。高德还有路况、步骤，这里不解析。
// Distance 单位米，Duration 单位秒，JSON 里就是字符串。
type directionDrivingResp struct {
	Status string `json:"status"`
	Info   string `json:"info"`
	Route  struct {
		Paths []struct {
			Distance string `json:"distance"`
			Duration string `json:"duration"`
		} `json:"paths"`
	} `json:"route"`
}
