package models

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"ginchat/utils"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

// Message 是一条聊天内容，也对应表 message。
// 当前收发不往这张表插行，私聊记录放在 Redis。表留着，是为了自动迁移时结构还在。
// 页面发来的 JSON 用 userId、TargetId、Type、Media、Content、url、amount。
// encoding/json 匹配字段时不区分大小写，所以 userId 能填进 UserId。
// 转发给对方时必须用页面原来的那一段 JSON。如果在服务器里重新 Marshal，
// 字段会变成 UserId，页面读 data.userId 就是空的。
type Message struct {
	Base
	UserId     int64  `gorm:"comment:发送者用户 ID"`
	TargetId   int64  `gorm:"comment:私聊时是对方用户 ID，群聊时是群 ID"`
	Type       int    `gorm:"comment:1 私聊，2 群聊，3 心跳"`
	Media      int    `gorm:"comment:1 文字，2 表情，3 语音，4 图片"`
	Content    string `gorm:"comment:文字内容"`
	CreateTime uint64 `gorm:"comment:页面带来的时间"`
	ReadTime   uint64 `gorm:"comment:预留的已读时间"`
	Pic        string `gorm:"comment:预留图片字段"`
	Url        string `gorm:"comment:语音或图片地址"`
	Desc       string `gorm:"comment:预留说明"`
	Amount     int    `gorm:"comment:语音时长，单位秒"`
}

func (Message) TableComment() string { return "消息表结构，私聊正文实际在 Redis" }

// TableName 固定成 message，避免被收成 messages。
func (table *Message) TableName() string {
	return "message"
}

// Node 是一个在线用户在本进程里的连接。
// 读协程负责收浏览器的 JSON，写协程负责把 DataQueue 里的字节写回浏览器，两边不要互相等待。
// clientMap 用读写锁。心跳时间用原子变量。socket 的写和关共用 writeMu。这三处保护的不是同一份数据，所以没有锁顺序问题。
type Node struct {
	Conn      *websocket.Conn // 和这个浏览器之间的长连接
	Addr      string          // 远端地址，会写进 Redis 的 online_ 值
	UserId    int64           // 这条连接属于谁，也是 clientMap 的键
	LoginTime uint64          // 这条连接建立时的 Unix 秒
	DataQueue chan []byte     // 待写给浏览器的原始 JSON，缓冲 50 条
	done      chan struct{}   // 关掉它，写协程就退出。只关一次
	once      sync.Once       // 保证 offline 只跑一遍。读、写、定时器都可能触发清理
	writeMu   sync.Mutex      // gorilla 的连接不能同时 Write 和 Close。写协程和 offline 共用这把锁
	// heartbeat 是最近一次心跳的 Unix 秒。读协程写，清理协程读，必须用原子操作。
	heartbeat uint64
}

var (
	// clientMap 只在本进程有效。一个用户 ID 对一条当前连接，后连上的会替换先连上的。
	clientMap = make(map[int64]*Node)
	// rwLocker 保护 clientMap。查在线用读锁，登记和删除用写锁。
	rwLocker sync.RWMutex
	// udpOn 为 1 时，dispatch 才会把消息再广播到局域网。StartUDP 成功后才置上。
	// 用原子变量：写发生在启动协程，读发生在之后的聊天协程。
	udpOn uint32
	// udpSend 是待广播的原始 JSON。缓冲 1024，满了就丢，避免拖住聊天协程。
	udpSend = make(chan []byte, 1024)
	// seenMsg 记下最近一分钟见过的消息摘要。本机 UDP 会把刚发出去的包收回来，用来挡第二次投递。
	seenMsg sync.Map
)

// MsgKey 拼出两个人私聊记录的 Redis 键，形如 msg_较小ID_较大ID。
// a、b 谁是发送者都行。先按数值从小到大排，双方打开会话读到的是同一个有序集合。
func MsgKey(a, b int64) string {
	if a > b {
		a, b = b, a
	}
	return fmt.Sprintf("msg_%d_%d", a, b)
}

// onlineKey 是 Redis 里的在线标记，形如 online_7。值是连接的远端地址。
func onlineKey(userId int64) string {
	return "online_" + strconv.FormatInt(userId, 10)
}

// onlineTTL 读取配置里的在线时长，单位是小时。没配或配成 0 时按 4 小时。
// 心跳每次都会用这个时长重新 SET，所以活跃用户的标记会一直续着。
func onlineTTL() time.Duration {
	hours := viper.GetInt("timeout.RedisOnlineTime")
	if hours <= 0 {
		hours = 4
	}
	return time.Duration(hours) * time.Hour
}

// Chat 把一次 HTTP 请求升级成 WebSocket。聊天页和 /user/sendUserMsg 都进这里。
// 查询参数 userId 是用户 ID，token 必须等于该用户 user_basic.identity。
// 缺参数、用户不存在、令牌为空或不一致，都回 401，不升级连接。
// 升级成功后：替换这个用户的旧连接，拉起读写两个协程，并在 Redis 写下在线标记。
func Chat(writer http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	idText := query.Get("userId")
	token := query.Get("token")
	userId, err := strconv.ParseInt(idText, 10, 64)
	if err != nil || userId <= 0 {
		http.Error(writer, "缺少 userId", http.StatusUnauthorized)
		return
	}
	user := FindByID(uint(userId))
	if user.ID == 0 || user.Identity == "" || user.Identity != token {
		http.Error(writer, "登录已失效", http.StatusUnauthorized)
		return
	}

	// CheckOrigin 固定放行，本机换个端口打开页面也能连。正式环境应按来源站点判断。
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		log.Println("websocket upgrade", err)
		return
	}

	now := uint64(time.Now().Unix())
	node := &Node{
		Conn:      conn,
		Addr:      conn.RemoteAddr().String(),
		UserId:    userId,
		LoginTime: now,
		DataQueue: make(chan []byte, 50),
		done:      make(chan struct{}),
	}
	atomic.StoreUint64(&node.heartbeat, now)
	// 先放进表，再让旧连接下线。旧连接的 offline 会核对「表里是不是自己」，
	// 此时表里已经是新连接，所以它只会关自己的 socket，不会把新连接删掉。
	rwLocker.Lock()
	if old, ok := clientMap[userId]; ok && old != node {
		go old.offline()
	}
	clientMap[userId] = node
	rwLocker.Unlock()

	go sendProc(node)
	go recvProc(node)
	SetUserOnlineInfo(onlineKey(userId), []byte(node.Addr), onlineTTL())
}

// sendProc 是这条连接的写协程。
// done 被关掉就退出；否则取出一条原始 JSON，按文本帧写给浏览器。写失败说明连接已经断了，随即做一次下线。
func sendProc(node *Node) {
	for {
		select {
		case <-node.done:
			return
		case data := <-node.DataQueue:
			if err := node.writeMessage(data); err != nil {
				log.Println("ws send", node.UserId, err)
				node.offline()
				return
			}
		}
	}
}

// recvProc 是这条连接的读协程。函数退出时一定会 offline，所以读出错、对端关闭都会清理。
// Type 为 3 只刷新心跳，不进入投递。其它消息把原始字节交给 dispatch，不在这里重新编码。
func recvProc(node *Node) {
	defer node.offline()
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			log.Println("ws read", node.UserId, err)
			return
		}
		var msg Message
		if err = json.Unmarshal(data, &msg); err != nil {
			log.Println("ws json", err)
			continue
		}
		if msg.Type == 3 {
			node.Heartbeat(uint64(time.Now().Unix()))
			continue
		}
		dispatch(data)
	}
}

// dispatch 按 Type 把原始 JSON 交给私聊或群发，然后再尝试 UDP 广播。
// 同一段字节如果从 UDP 又绕回来，seenMsg 里已经有它的摘要，直接丢掉，避免本机收到两次。
// 摘要保留一分钟后删除，免得这张表一直变大。
func dispatch(data []byte) {
	sum := md5.Sum(data)
	if _, loaded := seenMsg.LoadOrStore(sum, time.Now()); loaded {
		return
	}
	time.AfterFunc(time.Minute, func() { seenMsg.Delete(sum) })

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Println("dispatch json", err)
		return
	}
	msg.CreateTime = uint64(time.Now().Unix())
	switch msg.Type {
	case 1:
		sendMsg(msg.TargetId, data)
	case 2:
		sendGroupMsg(msg.TargetId, msg.UserId, data)
	default:
		log.Println("忽略未知消息类型", msg.Type)
	}
	broadMsg(data)
}

// sendGroupMsg 把一条群消息写给群里除发送者以外的成员。
// groupId 是群 ID，fromUser 是发送者。名单来自 contact 里 type=2 的行。
// 跳过发送者，是因为页面在发送前已经把这句话画在自己的屏幕上。
func sendGroupMsg(groupId int64, fromUser int64, msg []byte) {
	for _, id := range SearchUserByGroupId(uint(groupId)) {
		if int64(id) == fromUser {
			continue
		}
		sendMsg(int64(id), msg)
	}
}

// sendMsg 把一条消息交给指定用户。
// userId 在私聊里是对方，在群发里是每一个接收成员。msg 是浏览器原来的 JSON，不要改它。
// 本进程有这条连接，并且 Redis 里还有在线标记，才放进发送队列。
// 只有私聊（Type=1）才按两人 ID 写进 Redis。群消息只负责送达，不混进单聊历史。
// 有序集合的分数用当前毫秒。用「当前条数 + 1」当分数时，并发插入会撞分，所以不用条数。
func sendMsg(userId int64, msg []byte) {
	if userId <= 0 {
		return
	}
	var header Message
	if err := json.Unmarshal(msg, &header); err != nil {
		log.Println("sendMsg json", err)
		return
	}

	rwLocker.RLock()
	node, onlineHere := clientMap[userId]
	rwLocker.RUnlock()
	if onlineHere && redisOnline(userId) {
		node.enqueue(msg)
	}
	// 群消息只负责投递。私聊才按两个人的 ID 存进 Redis，避免群记录混进单聊历史。
	if header.Type != 1 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := utils.Red.ZAdd(ctx, MsgKey(header.UserId, userId), &redis.Z{
		Score:  float64(time.Now().UnixMilli()),
		Member: string(msg),
	}).Err()
	if err != nil {
		log.Println("保存聊天记录", err)
	}
}

// enqueue 把一条消息放进写队列。
// 连接已经在下线，或者队列里已经有 50 条还没写出去，这条就丢掉并打日志，调用方不会被堵住。
func (node *Node) enqueue(msg []byte) {
	select {
	case <-node.done:
	case node.DataQueue <- msg:
	default:
		log.Println("发送队列已满", node.UserId)
	}
}

// redisOnline 看 online_用户ID 这个键还在不在。
// 键被心跳续着就视为在线。查询失败或键不存在都当离线，消息仍然可以稍后从历史接口读到。
func redisOnline(userId int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	n, err := utils.Red.Exists(ctx, onlineKey(userId)).Result()
	return err == nil && n > 0
}

// RedisMsg 按时间取出两个人之间的一段私聊记录。
// userIdA、userIdB 不分发送者和接收者，MsgKey 会自己排序。
// start、end 是有序集合的下标，从 0 开始，不是用户 ID。页面第一次打开会话时传 0 和 9。
// isRev 为 false 时从新到旧，和页面第一次打开会话的方向一致；为 true 时从旧到新。
// 返回的每一项都是当初浏览器发出的 JSON 字符串。
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	key := MsgKey(userIdA, userIdB)
	var (
		rows []string
		err  error
	)
	if isRev {
		rows, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rows, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// writeMessage 是这条连接唯一的写入口。和 offline 里的 Close 共用 writeMu。
// gorilla/websocket 规定同一条连接不能并发写，也不能一边写一边关。
func (node *Node) writeMessage(data []byte) error {
	node.writeMu.Lock()
	defer node.writeMu.Unlock()
	return node.Conn.WriteMessage(websocket.TextMessage, data)
}

// Heartbeat 记下这条连接还活着。
// currentTime 是当前的 Unix 秒。清理任务在另一条协程里读它，所以用原子写。
// 同时把 Redis 的 online_ 键按配置时长重新写一遍，等于给在线标记续期。
func (node *Node) Heartbeat(currentTime uint64) {
	atomic.StoreUint64(&node.heartbeat, currentTime)
	SetUserOnlineInfo(onlineKey(node.UserId), []byte(node.Addr), onlineTTL())
}

// offline 关掉这条连接，并且只做一次。
// 先从 clientMap 删掉自己，但只在表里仍是这条连接时才删，避免误删后来顶上来的新连接。
// 然后关掉 done，写协程会退出；删掉 Redis 在线标记；最后关闭 socket。
func (node *Node) offline() {
	node.once.Do(func() {
		rwLocker.Lock()
		if cur, ok := clientMap[node.UserId]; ok && cur == node {
			delete(clientMap, node.UserId)
		}
		rwLocker.Unlock()
		close(node.done)
		if utils.Red != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			utils.Red.Del(ctx, onlineKey(node.UserId))
		}
		node.writeMu.Lock()
		_ = node.Conn.Close()
		node.writeMu.Unlock()
	})
}

// CleanConnection 关掉心跳超时的连接。param 没有使用，定时器的签名要求带上它。
// 先在读锁里把超时的连接抄出来，再逐个 offline。不要拿着锁去关 socket，关的时候还要再拿写锁。
// 返回 true，定时器会按间隔继续跑。函数内部如果 panic，被 recover 吃掉，避免把清理协程打停。
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			log.Println("cleanConnection", r)
		}
	}()
	now := uint64(time.Now().Unix())
	rwLocker.RLock()
	expired := make([]*Node, 0)
	for _, node := range clientMap {
		if node.IsHeartbeatTimeOut(now) {
			expired = append(expired, node)
		}
	}
	rwLocker.RUnlock()
	for _, node := range expired {
		log.Println("心跳超时，关闭连接", node.UserId)
		node.offline()
	}
	return result
}

// IsHeartbeatTimeOut 判断这条连接是不是太久没有心跳。
// currentTime 和心跳时间都是 Unix 秒。配置 timeout.HeartbeatMaxTime 也是秒，直接相加。
// 配置缺失或为 0 时返回 false，不把所有人踢下线。
// 心跳由读协程更新，本函数由定时清理协程调用，两边用原子读，避免数据竞争。
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) bool {
	maxWait := viper.GetUint64("timeout.HeartbeatMaxTime")
	if maxWait == 0 {
		return false
	}
	return atomic.LoadUint64(&node.heartbeat)+maxWait <= currentTime
}

// StartUDP 监听配置里的 UDP 端口，并把以后的业务消息广播到 255.255.255.255。
// 另一台机器上的 GinChat 如果也听这个端口，就能收到并投递给自己的在线用户。
// 监听或拨号失败只记日志，然后返回。本机两个浏览器的消息走 dispatch，不经过这里。
func StartUDP() {
	port := viper.GetInt("port.udp")
	if port <= 0 {
		log.Println("未配置 port.udp，跳过局域网广播")
		return
	}
	recv, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: port})
	if err != nil {
		log.Println("udp 监听失败，单机聊天仍可用:", err)
		return
	}
	send, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.IPv4bcast, Port: port})
	if err != nil {
		log.Println("udp 广播不可用，单机聊天仍可用:", err)
		_ = recv.Close()
		return
	}
	atomic.StoreUint32(&udpOn, 1)
	go udpSendLoop(send)
	go udpRecvLoop(recv)
	log.Println("udp broadcast on", port)
}

// broadMsg 把原始 JSON 丢进 UDP 发送队列。UDP 没启动，或队列已满，就什么都不做。
func broadMsg(data []byte) {
	if atomic.LoadUint32(&udpOn) == 0 {
		return
	}
	select {
	case udpSend <- data:
	default:
	}
}

// udpSendLoop 把队列里的字节写出广播连接。写失败就退出，本机聊天不受影响。
func udpSendLoop(conn *net.UDPConn) {
	defer conn.Close()
	for data := range udpSend {
		if _, err := conn.Write(data); err != nil {
			log.Println("udp write", err)
			return
		}
	}
}

// udpRecvLoop 读到局域网里的一份 JSON 后交给 dispatch。
// 先把这一小段复制出来再交给后面，因为 buf 下一次读取会被覆盖。
// 自己刚广播出去的包也会被这里读到，dispatch 里的摘要表会把它丢掉。
func udpRecvLoop(conn *net.UDPConn) {
	defer conn.Close()
	buf := make([]byte, 4096)
	for {
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("udp read", err)
			return
		}
		payload := append([]byte(nil), buf[:n]...)
		dispatch(payload)
	}
}

// MarshalBinary 让 Message 满足 Redis 客户端对结构体成员的编码接口。
// 当前私聊存的是浏览器原始 JSON 字符串，不走这个方法。以后如果要直接把结构体放进有序集合，才会用到它。
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}
