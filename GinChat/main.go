// GinChat 是一个单进程的即时聊天。
// 浏览器只访问这一个进程：页面和表单走 Gin，聊天走 WebSocket。
// 用户、好友、群在 MySQL；在线状态和聊天记录在 Redis。
//
// @title       GinChat
// @version     1.0
// @description 单机即时聊天。登录、好友、群走 HTTP，消息走 WebSocket。
// @host        127.0.0.1:8080
// @BasePath    /
package main

import (
	"ginchat/models"
	"ginchat/router"
	"ginchat/utils"
	"log"
	"time"

	"github.com/spf13/viper"
)

// main 按固定顺序把进程拉起来。任何一步失败都直接退出，避免后面第一次请求才发现库没连上。
// 顺序不能倒：建表要用已经打开的 MySQL，心跳清理和 UDP 都要先读到配置。
func main() {
	// 工作目录必须是 GinChat 根目录，config/app.yml 是相对路径。
	if err := utils.InitConfig(); err != nil {
		log.Fatal(err)
	}
	// 库不存在会先 CREATE DATABASE。账号密码错在这里就退出。
	if err := utils.InitMySQL(); err != nil {
		log.Fatal(err)
	}
	// 按模型建表、补列。不会删列，也不会改已经写入的数据。
	if err := utils.AutoMigrate(
		&models.UserBasic{},
		&models.Message{},
		&models.Contact{},
		&models.Community{},
		&models.GroupBasic{},
	); err != nil {
		log.Fatal(err)
	}
	// 在线标记和私聊记录都在 Redis。Ping 失败就退出。
	if err := utils.InitRedis(); err != nil {
		log.Fatal(err)
	}
	// 局域网广播失败只打日志。同一台机器上的两个浏览器不依赖 UDP。
	models.StartUDP()
	// 延迟 DelayHeartbeat 秒后开始扫连接，之后每 HeartbeatHz 秒再扫一次。
	// CleanConnection 一直返回 true，所以这个定时器不会自己停。
	utils.Timer(
		time.Duration(viper.GetInt("timeout.DelayHeartbeat"))*time.Second,
		time.Duration(viper.GetInt("timeout.HeartbeatHz"))*time.Second,
		models.CleanConnection,
		nil,
	)

	// port.server 形如 ":8080"，冒号不能省，否则 Gin 会把它当成主机名。
	addr := viper.GetString("port.server")
	log.Println("GinChat listening", addr)
	if err := router.Router().Run(addr); err != nil {
		log.Fatal(err)
	}
}
