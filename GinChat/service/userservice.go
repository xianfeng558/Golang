// Package service 是 HTTP 层。
// 它从 Gin 的请求里取出表单和查询参数，调用 models，再按页面习惯写出 JSON。
// 用户相关接口用小写的 code、message、data。好友、群、上传、历史消息用大写的 Code、Msg、Rows、Data。
// 页面的 util.post 发送的是 application/x-www-form-urlencoded，所以这里用 FormValue / PostForm，不读 JSON 正文。
package service

import (
	"ginchat/models"
	"ginchat/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// writeUser 写登录、注册、改资料用的小写 JSON。
// code 为 0 表示成功，-1 表示失败。HTTP 状态固定 200，页面只看 code 和 message。
// data 是用户对象。密码和盐因为 json:"-" 不会出现在响应里。
func writeUser(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

// GetUserList 返回全部未删除用户。
// 聊天页不调用它。成功时 data 是用户数组，失败时 code 为 -1。
// @Summary  所有用户
// @Tags     用户模块
// @Accept   x-www-form-urlencoded
// @Success  200 {object} map[string]interface{}
// @Router   /user/getUserList [post]
func GetUserList(c *gin.Context) {
	data, err := models.GetUserList()
	if err != nil {
		writeUser(c, -1, "获取用户列表失败", nil)
		return
	}
	writeUser(c, 0, "获取用户列表成功", data)
}

// CreateUser 处理注册。
// 必填字段是 name、password。确认密码优先读 Identity，因为注册页输入框绑的是这个名字；没有再读 repassword。
// 用户名为空、两次密码不一致、名字已经存在，都返回 code=-1，不插入。
// 通过后生成 6 位盐，把 Md5(明文+盐) 写入 pass_word，再插入。响应里的 data.ID 是新用户的主键。
// @Summary  新增用户
// @Tags     用户模块
// @Accept   x-www-form-urlencoded
// @Param    name       formData string true  "用户名"
// @Param    password   formData string true  "密码"
// @Param    Identity   formData string false "确认密码，注册页使用这个字段"
// @Param    repassword formData string false "确认密码"
// @Success  200 {object} map[string]interface{}
// @Router   /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{Name: c.Request.FormValue("name")}
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("Identity")
	if repassword == "" {
		repassword = c.Request.FormValue("repassword")
	}
	if user.Name == "" || password == "" || repassword == "" {
		writeUser(c, -1, "用户名或密码不能为空！", nil)
		return
	}
	if password != repassword {
		writeUser(c, -1, "两次密码不一致！", nil)
		return
	}
	if models.FindUserByName(user.Name).ID != 0 {
		writeUser(c, -1, "用户名已注册！", nil)
		return
	}
	user.Salt = models.NewSalt()
	user.PassWord = utils.MakePassword(password, user.Salt)
	now := time.Now()
	user.LoginTime = now
	user.LoginOutTime = now
	user.HeartbeatTime = now
	if err := models.CreateUser(&user); err != nil {
		writeUser(c, -1, "新增用户失败", nil)
		return
	}
	writeUser(c, 0, "新增用户成功！", user)
}

// FindUserByNameAndPwd 是登录。
// 先按 name 找人，再用该用户的盐校验 password。校验通过后签发新的 Identity，旧令牌立即失效。
// 成功时 data.ID 给页面当 userId，data.Identity 给页面当 token。用户不存在或密码错误都是 code=-1。
// @Summary  登录
// @Tags     用户模块
// @Accept   x-www-form-urlencoded
// @Param    name     formData string true "用户名"
// @Param    password formData string true "密码"
// @Success  200 {object} map[string]interface{}
// @Router   /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	user := models.FindUserByName(name)
	if user.ID == 0 {
		writeUser(c, -1, "该用户不存在", nil)
		return
	}
	if !utils.ValidPassword(password, user.Salt, user.PassWord) {
		writeUser(c, -1, "密码不正确", nil)
		return
	}
	if err := models.IssueToken(&user); err != nil {
		writeUser(c, -1, "登录失败", nil)
		return
	}
	writeUser(c, 0, "登录成功", user)
}

// DeleteUser 按表单里的 id 软删除用户。
// id 缺失或不是数字时会变成 0，模型层会拒绝，这里返回「删除用户失败」。
// 软删除后这个人不再出现在列表和登录里，历史关系行还在。
// @Summary  删除用户
// @Tags     用户模块
// @Accept   x-www-form-urlencoded
// @Param    id formData string true "用户 ID"
// @Success  200 {object} map[string]interface{}
// @Router   /user/deleteUser [post]
func DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("id"))
	user := models.UserBasic{}
	user.ID = uint(id)
	if err := models.DeleteUser(user); err != nil {
		writeUser(c, -1, "删除用户失败", nil)
		return
	}
	writeUser(c, 0, "删除用户成功！", user)
}

// UpdateUser 修改昵称、头像、手机、邮箱或密码。
// 页面改资料只提交 id、name、icon。icon 写入 Avatar。手机和邮箱可以空着，空值不会覆盖原列。
// phone、email 一旦有值，就要过 govalidator：手机是 11 位，邮箱要像邮箱。校验失败返回「修改参数不匹配」。
// 成功后重新查出这个用户再返回，让响应里的昵称和头像是库里的新值。
// @Summary  修改用户
// @Tags     用户模块
// @Accept   x-www-form-urlencoded
// @Param    id       formData string true  "用户 ID"
// @Param    name     formData string false "昵称"
// @Param    password formData string false "新密码，不填则不改"
// @Param    phone    formData string false "手机"
// @Param    email    formData string false "邮箱"
// @Param    icon     formData string false "头像地址"
// @Success  200 {object} map[string]interface{}
// @Router   /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	user := models.UserBasic{
		Name:     c.PostForm("name"),
		PassWord: c.PostForm("password"),
		Phone:    c.PostForm("phone"),
		Avatar:   c.PostForm("icon"),
		Email:    c.PostForm("email"),
	}
	user.ID = uint(id)
	if _, err := govalidator.ValidateStruct(user); err != nil {
		writeUser(c, -1, "修改参数不匹配！", nil)
		return
	}
	if err := models.UpdateProfile(user); err != nil {
		writeUser(c, -1, "修改用户失败", nil)
		return
	}
	writeUser(c, 0, "修改用户成功！", models.FindByID(user.ID))
}

// upGrader 只给演示用的 /user/sendMsg。聊天页的升级在 models.Chat，不走这个变量。
// CheckOrigin 放行任意页面来源，和学习环境里的聊天连接同一尺度。
var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SendMsg 把 Redis 频道 websocket 里的新消息推给这条连接。
// 它演示发布订阅，不校验登录，也不进入聊天的投递逻辑。聊天页连接的是 /chat。
// 连接断开或上下文取消时，订阅和 socket 都会关掉，不会把进程打退出。
// @Summary  订阅 Redis 消息
// @Tags     聊天
// @Success  101 {string} string "切换到 WebSocket"
// @Router   /user/sendMsg [get]
func SendMsg(c *gin.Context) {
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("upgrade", err)
		return
	}
	defer ws.Close()
	sub := utils.Red.Subscribe(c.Request.Context(), utils.PublishKey)
	defer sub.Close()
	for {
		msg, err := sub.ReceiveMessage(c.Request.Context())
		if err != nil {
			log.Println("subscribe", err)
			return
		}
		line := "[" + time.Now().Format("2006-01-02 15:04:05") + "] " + msg.Payload
		if err = ws.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			log.Println("ws write", err)
			return
		}
	}
}

// RedisMsg 读取两个人的私聊历史。
// userIdA、userIdB 是两个人的 ID，不分谁发给谁。start、end 是有序集合下标，页面第一次传 0 和 9。
// isRev 为 false 时从新到旧。结果在 Rows，每一项是页面当初发出的 JSON 字符串；Total 是这一段的条数。
// @Summary  历史消息
// @Tags     聊天
// @Accept   x-www-form-urlencoded
// @Param    userIdA formData int  true  "其中一个用户"
// @Param    userIdB formData int  true  "另一个用户"
// @Param    start   formData int  false "起始下标"
// @Param    end     formData int  false "结束下标"
// @Param    isRev   formData bool false "true 时从旧到新"
// @Success  200 {object} utils.H
// @Router   /user/redisMsg [post]
func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.ParseInt(c.PostForm("userIdA"), 10, 64)
	userIdB, _ := strconv.ParseInt(c.PostForm("userIdB"), 10, 64)
	start, _ := strconv.ParseInt(c.PostForm("start"), 10, 64)
	end, _ := strconv.ParseInt(c.PostForm("end"), 10, 64)
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	rows, err := models.RedisMsg(userIdA, userIdB, start, end, isRev)
	if err != nil {
		utils.RespFail(c.Writer, "读取消息失败")
		return
	}
	utils.RespOKList(c.Writer, rows, len(rows))
}

// SendUserMsg 和 /chat 一样，升级为聊天 WebSocket。
// 查询参数同样要有 userId 和 token。保留这个地址，是因为旧页面可能还指着它。
// @Summary  聊天 WebSocket
// @Tags     聊天
// @Param    userId query string true "用户 ID"
// @Param    token  query string true "登录令牌"
// @Router   /user/sendUserMsg [get]
func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

// SearchFriends 按 userId 返回好友列表。
// 数据在 Rows，每一项是用户对象。Total 是好友人数。没有好友时 Rows 是空数组，Code 仍是 0。
// @Summary  好友列表
// @Tags     好友和群
// @Accept   x-www-form-urlencoded
// @Param    userId formData int true "用户 ID"
// @Success  200 {object} utils.H
// @Router   /searchFriends [post]
func SearchFriends(c *gin.Context) {
	id, _ := strconv.Atoi(c.Request.FormValue("userId"))
	users := models.SearchFriend(uint(id))
	utils.RespOKList(c.Writer, users, len(users))
}

// AddFriend 按对方用户名加好友。
// userId 是自己的数字 ID，targetName 是对方登录名。成功时 Code 为 0，失败时 Code 为 -1，原因在 Msg。
// @Summary  添加好友
// @Tags     好友和群
// @Accept   x-www-form-urlencoded
// @Param    userId     formData int    true "自己的 ID"
// @Param    targetName formData string true "对方用户名"
// @Success  200 {object} utils.H
// @Router   /contact/addfriend [post]
func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	code, msg := models.AddFriend(uint(userId), c.Request.FormValue("targetName"))
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
		return
	}
	utils.RespFail(c.Writer, msg)
}

// CreateCommunity 建群。
// ownerId 是群主的用户 ID，name 必填。icon 是上传接口返回的头像路径，写入群的 Img。desc 是简介。
// 成功时群主自己的群列表里会立刻出现这个群。失败原因在 Msg，例如群名为空。
// @Summary  创建群
// @Tags     好友和群
// @Accept   x-www-form-urlencoded
// @Param    ownerId formData int    true  "群主 ID"
// @Param    name    formData string true  "群名称"
// @Param    icon    formData string false "头像地址"
// @Param    desc    formData string false "简介"
// @Success  200 {object} utils.H
// @Router   /contact/createCommunity [post]
func CreateCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	community := models.Community{
		OwnerId: uint(ownerId),
		Name:    c.Request.FormValue("name"),
		Img:     c.Request.FormValue("icon"),
		Desc:    c.Request.FormValue("desc"),
	}
	code, msg := models.CreateCommunity(community)
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
		return
	}
	utils.RespFail(c.Writer, msg)
}

// LoadCommunity 返回 ownerId 这个人加入的群，包括他自己建的。
// 数据在 Rows。一个人都没群时 Rows 是空数组，Total 是「查询成功」，Code 仍是 0。
// @Summary  群列表
// @Tags     好友和群
// @Accept   x-www-form-urlencoded
// @Param    ownerId formData int true "用户 ID"
// @Success  200 {object} utils.H
// @Router   /contact/loadcommunity [post]
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.Request.FormValue("ownerId"))
	data, msg := models.LoadCommunity(uint(ownerId))
	utils.RespList(c.Writer, 0, data, msg)
}

// JoinGroups 把 userId 加进一个群。
// comId 可以是群的数字 ID，也可以是群名称，页面输入框两种都接受。
// 已经在群里、群不存在，都返回 Code=-1 和对应的 Msg。
// @Summary  加入群
// @Tags     好友和群
// @Accept   x-www-form-urlencoded
// @Param    userId formData int    true "用户 ID"
// @Param    comId  formData string true "群 ID 或群名称"
// @Success  200 {object} utils.H
// @Router   /contact/joinGroup [post]
func JoinGroups(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	code, msg := models.JoinGroup(uint(userId), c.Request.FormValue("comId"))
	if code == 0 {
		utils.RespOK(c.Writer, code, msg)
		return
	}
	utils.RespFail(c.Writer, msg)
}

// FindByID 按 userId 取一个人的资料，结果在 Data。
// 聊天页收到一条消息后，用发送者 ID 调它，补上头像和昵称。找不到人时 Code 为 -1。
// @Summary  查询用户
// @Tags     用户模块
// @Accept   x-www-form-urlencoded
// @Param    userId formData int true "用户 ID"
// @Success  200 {object} utils.H
// @Router   /user/find [post]
func FindByID(c *gin.Context) {
	userId, _ := strconv.Atoi(c.Request.FormValue("userId"))
	data := models.FindByID(uint(userId))
	if data.ID == 0 {
		utils.RespFail(c.Writer, "没有找到此用户")
		return
	}
	utils.RespOK(c.Writer, data, "ok")
}
