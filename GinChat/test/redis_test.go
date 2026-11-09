package chattest

import (
	"context"
	"ginchat/utils"
	"os"
	"testing"
	"time"
)

// TestPublish 往演示频道 websocket 发一条带时间的字符串。
// 平时跳过，避免 go test 每次都改 Redis。需要看发布订阅时再开：
//
//	set GINCHAT_REDIS_TEST=1
//	go test ./test -run TestPublish -v
func TestPublish(t *testing.T) {
	if os.Getenv("GINCHAT_REDIS_TEST") == "" {
		t.Skip("设置 GINCHAT_REDIS_TEST=1 才向 Redis 发布一条测试消息")
	}
	if err := utils.InitConfig(); err != nil {
		t.Fatal(err)
	}
	if err := utils.InitRedis(); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := utils.Publish(ctx, utils.PublishKey, "当前时间: "+time.Now().Format("15:04:05")); err != nil {
		t.Fatal(err)
	}
}
