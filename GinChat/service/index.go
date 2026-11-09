package service

import (
	"ginchat/models"
	"net/http"
	"strconv"
	"text/template"

	"github.com/gin-gonic/gin"
)

// GetIndex 打开登录页。/ 和 /index 都进这里。
// 模板在项目根目录的 index.html，它用 {{template "/chat/head.shtml"}} 引用公共脚本，
// 所以解析时必须把 views/chat/head.html 一起带上。传给模板的数据是 nil，页面上不再印出多余文字。
// @Tags     首页
// @Success  200 {string} string "登录页"
// @Router   /index [get]
func GetIndex(c *gin.Context) {
	render(c, nil, "index.html", "views/chat/head.html")
}

// ToRegister 打开注册页。注册页不嵌别的模板，只解析 views/user/register.html。
// @Tags    首页
// @Success 200 {string} string "注册页"
// @Router  /toRegister [get]
func ToRegister(c *gin.Context) {
	render(c, nil, "views/user/register.html")
}

// ToChat 打开聊天页。
// 查询参数 userId、token 放进一个 UserBasic，模板如果要印用户信息可以用它们。
// 真正连 WebSocket 的是 foot.html 里的脚本，它自己再从地址栏读这两个参数。
// 下面列出的 html 都带 {{define}}，必须一次解析，否则页面里的 {{template}} 找不到对应的块。
// @Tags     首页
// @Param    userId query string true "用户 ID"
// @Param    token  query string true "登录返回的 Identity"
// @Success  200 {string} string "聊天页"
// @Router   /toChat [get]
func ToChat(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Query("userId"))
	user := models.UserBasic{Identity: c.Query("token")}
	user.ID = uint(userId)
	render(c, user,
		"views/chat/index.html",
		"views/chat/head.html",
		"views/chat/foot.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/createcom.html",
		"views/chat/userinfo.html",
		"views/chat/main.html",
	)
}

// Chat 是聊天页使用的 WebSocket 入口。
// 这里只把响应和请求交给 models.Chat。令牌核对、升级、读写协程都在模型里完成。
// @Tags    聊天
// @Param   userId query string true "用户 ID"
// @Param   token  query string true "登录令牌"
// @Success 101 {string} string "切换到 WebSocket"
// @Router  /chat [get]
func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

// render 解析一组模板并把第一份当成入口执行。
// files[0] 是要输出的那一页，后面的文件提供它引用的 {{define}}。
// 解析或执行失败时回 500 和错误文本，不让一次坏页面把进程打退出。
func render(c *gin.Context, data interface{}, files ...string) {
	tpl, err := template.ParseFiles(files...)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if err = tpl.Execute(c.Writer, data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
