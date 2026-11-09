// Package router 把 URL 交给 service。
// 页面、用户、好友、群、上传、WebSocket 都在这一个 Engine 上。
package router

import (
	"ginchat/service"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ginchat/docs" // 注册 swag 生成的接口说明，浏览器打开 /swagger/index.html
)

// Router 返回已经挂好路由的 Gin。
// 静态文件来自 asset/。views 下的模板在这里预加载，登录页 index.html 不在 views 里，由 service 自己解析。
// 用户接口大多是 POST，因为页面的 util.post 固定用表单，不用 JSON 正文。
func Router() *gin.Engine {
	r := gin.Default()
	// 接口说明页：http://127.0.0.1:8080/swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 页面上的 css、js、图片、上传结果都从 /asset 取。
	r.Static("/asset", "asset/")
	r.StaticFile("/favicon.ico", "asset/images/favicon.ico")
	r.LoadHTMLGlob("views/**/*")

	// 页面。/ 和 /index 都是登录页。
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	// 聊天长连接。查询参数要带 userId 和 token。
	r.GET("/chat", service.Chat)
	// 好友列表。字段名是 userId，结果在 Rows。
	r.POST("/searchFriends", service.SearchFriends)

	// 用户。这组响应是小写的 code、message、data。
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	// 按用户 ID 取资料。聊天页收到消息后用它补头像和昵称，结果在 Data。
	r.POST("/user/find", service.FindByID)
	// 这条 WebSocket 只订阅 Redis 的 websocket 频道，用来看发布订阅，聊天页不用它。
	r.GET("/user/sendMsg", service.SendMsg)
	// 和 /chat 同一条升级逻辑，保留给旧地址。
	r.GET("/user/sendUserMsg", service.SendUserMsg)
	// 私聊历史。userIdA、userIdB 是两个人，start、end 是下标。
	r.POST("/user/redisMsg", service.RedisMsg)

	// 好友和群。这组响应是大写的 Code、Msg、Rows。
	r.POST("/contact/addfriend", service.AddFriend)
	r.POST("/contact/createCommunity", service.CreateCommunity)
	r.POST("/contact/loadcommunity", service.LoadCommunity)
	r.POST("/contact/joinGroup", service.JoinGroups)
	// 图片和语音。表单字段名必须是 file。
	r.POST("/attach/upload", service.Upload)
	return r
}
