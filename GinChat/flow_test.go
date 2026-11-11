package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"ginchat/models"
	"ginchat/router"
	"ginchat/utils"

	"github.com/gorilla/websocket"
)

// TestChatFlow 在本机 MySQL 和 Redis 上把一条聊天链路跑通。
// 顺序是：打开登录页和注册页，注册两个人，拒绝重复注册和错误密码，登录拿到不同令牌，
// 改资料，加好友，建群，按群名加群，私聊并写进 Redis，群消息只投递给成员，上传文件，错误令牌被拒绝。
// 它自己起一个临时 HTTP 服务，不占用配置里的 8080。
func TestChatFlow(t *testing.T) {
	if err := utils.InitConfig(); err != nil {
		t.Fatal(err)
	}
	if err := utils.InitMySQL(); err != nil {
		t.Fatal(err)
	}
	if err := utils.AutoMigrate(
		&models.UserBasic{},
		&models.Message{},
		&models.Contact{},
		&models.Community{},
		&models.GroupBasic{},
		&models.FriendApply{},
	); err != nil {
		t.Fatal(err)
	}
	if err := utils.InitRedis(); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(router.Router())
	defer srv.Close()
	base := srv.URL

	for _, page := range []struct {
		path string
		text string
	}{
		{"/", "登录"},
		{"/index", "登录"},
		{"/toRegister", "注册"},
	} {
		body := getOK(t, base+page.path)
		if !strings.Contains(body, page.text) {
			t.Fatalf("%s 页面里没有 %q", page.path, page.text)
		}
	}

	stamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	nameA := "gina_" + stamp
	nameB := "ginb_" + stamp
	createUser(t, base, nameA, "pass123")
	createUser(t, base, nameB, "pass123")
	again := postForm(t, base+"/user/createUser", url.Values{
		"name": {nameA}, "password": {"pass123"}, "Identity": {"pass123"},
	})
	if codeOf(again) != -1 {
		t.Fatalf("重复注册应失败: %#v", again)
	}

	bad := postForm(t, base+"/user/findUserByNameAndPwd", url.Values{
		"name": {nameA}, "password": {"wrong"},
	})
	if codeOf(bad) != -1 {
		t.Fatalf("错误密码应失败: %#v", bad)
	}

	loginA := login(t, base, nameA, "pass123")
	loginB := login(t, base, nameB, "pass123")
	if loginA.token == "" || loginA.token == loginB.token {
		t.Fatalf("两个用户的令牌应该不同: %s %s", loginA.token, loginB.token)
	}

	chatPage := getOK(t, base+"/toChat?userId="+loginA.id+"&token="+loginA.token)
	if !strings.Contains(chatPage, "initwebsocket") {
		t.Fatal("聊天页没有连接 WebSocket 的脚本")
	}

	updated := postForm(t, base+"/user/updateUser", url.Values{
		"id": {loginA.id}, "name": {nameA}, "icon": {"/asset/images/avatar0.png"},
	})
	if codeOf(updated) != 0 {
		t.Fatalf("修改资料失败: %#v", updated)
	}

	if postCode(t, base+"/contact/addfriend", url.Values{
		"userId": {loginA.id}, "targetName": {nameB},
	}) != 0 {
		t.Fatal("加好友失败")
	}
	nameC := "ginc_" + stamp
	createUser(t, base, nameC, "pass123")
	loginC := login(t, base, nameC, "pass123")
	if postCode(t, base+"/contact/apply", url.Values{
		"userId": {loginA.id}, "targetName": {nameC}, "remark": {"一起学习"},
	}) != 0 {
		t.Fatal("好友申请失败")
	}
	if postCode(t, base+"/contact/apply", url.Values{
		"userId": {loginA.id}, "targetName": {nameB},
	}) != -1 {
		t.Fatal("已经是好友时不应再申请成功")
	}
	listed := postBody(t, base+"/contact/applyList", url.Values{"userId": {loginC.id}})
	if !strings.Contains(listed, "一起学习") {
		t.Fatalf("申请列表里没有附言: %s", listed)
	}
	applyID := findJSONID(t, listed)
	if postCode(t, base+"/contact/applyReply", url.Values{
		"userId": {loginC.id}, "applyId": {applyID}, "accept": {"1"},
	}) != 0 {
		t.Fatal("同意好友申请失败")
	}
	friends := postBody(t, base+"/searchFriends", url.Values{"userId": {loginA.id}})
	if !strings.Contains(friends, nameB) || !strings.Contains(friends, nameC) {
		t.Fatalf("好友列表应同时有 %s 和 %s: %s", nameB, nameC, friends)
	}

	groupName := "群" + stamp
	if postCode(t, base+"/contact/createCommunity", url.Values{
		"ownerId": {loginA.id}, "name": {groupName}, "desc": {"测试群"},
	}) != 0 {
		t.Fatal("建群失败")
	}
	groups := postBody(t, base+"/contact/loadcommunity", url.Values{"ownerId": {loginA.id}})
	groupID := findIDByName(t, groups, groupName)
	if postCode(t, base+"/contact/joinGroup", url.Values{
		"userId": {loginB.id}, "comId": {groupName},
	}) != 0 {
		t.Fatal("按群名加群失败")
	}

	connA := dialChat(t, base, loginA)
	connB := dialChat(t, base, loginB)
	defer connA.Close()
	defer connB.Close()

	text := "hello-" + stamp
	_ = connB.SetReadDeadline(time.Now().Add(3 * time.Second))
	sendChat(t, connA, loginA.id, loginB.id, 1, text)
	if got := readChat(t, connB); !strings.Contains(got, text) {
		t.Fatalf("对方没有收到私聊: %s", got)
	}

	history := ""
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		history = postBody(t, base+"/user/redisMsg", url.Values{
			"userIdA": {loginA.id}, "userIdB": {loginB.id}, "start": {"0"}, "end": {"9"}, "isRev": {"false"},
		})
		if strings.Contains(history, text) {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !strings.Contains(history, text) {
		t.Fatalf("Redis 历史里没有这条私聊: %s", history)
	}

	_ = connB.SetReadDeadline(time.Now().Add(3 * time.Second))
	sendChat(t, connA, loginA.id, groupID, 2, "group-"+stamp)
	if got := readChat(t, connB); !strings.Contains(got, "group-"+stamp) {
		t.Fatalf("群成员没有收到群消息: %s", got)
	}

	uploadURL := uploadTinyPNG(t, base+"/attach/upload")
	if !strings.HasPrefix(uploadURL, "/asset/upload/") {
		t.Fatalf("上传地址不对: %s", uploadURL)
	}
	_ = os.Remove(filepath.Join("asset", "upload", filepath.Base(uploadURL)))

	wsURL := strings.Replace(base, "http", "ws", 1) + "/chat?userId=" + loginA.id + "&token=bad"
	badConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		badConn.Close()
		t.Fatal("错误令牌不应该能升级 WebSocket")
	}
}

type session struct {
	id    string
	token string
}

func createUser(t *testing.T, base, name, password string) {
	t.Helper()
	body := postForm(t, base+"/user/createUser", url.Values{
		"name": {name}, "password": {password}, "Identity": {password},
	})
	if codeOf(body) != 0 {
		t.Fatalf("注册 %s 失败: %#v", name, body)
	}
}

func login(t *testing.T, base, name, password string) session {
	t.Helper()
	body := postForm(t, base+"/user/findUserByNameAndPwd", url.Values{
		"name": {name}, "password": {password},
	})
	if codeOf(body) != 0 {
		t.Fatalf("登录 %s 失败: %#v", name, body)
	}
	data, _ := body["data"].(map[string]interface{})
	id, _ := data["ID"].(float64)
	token, _ := data["Identity"].(string)
	if id == 0 || token == "" {
		t.Fatalf("登录结果缺少 ID 或令牌: %#v", body)
	}
	return session{id: strconv.FormatInt(int64(id), 10), token: token}
}

func postForm(t *testing.T, rawURL string, form url.Values) map[string]interface{} {
	t.Helper()
	resp, err := http.PostForm(rawURL, form)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]interface{}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

func postBody(t *testing.T, rawURL string, form url.Values) string {
	t.Helper()
	resp, err := http.PostForm(rawURL, form)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func postCode(t *testing.T, rawURL string, form url.Values) int {
	t.Helper()
	body := postForm(t, rawURL, form)
	code, ok := body["Code"].(float64)
	if !ok {
		t.Fatalf("响应没有 Code: %#v", body)
	}
	return int(code)
}

func codeOf(body map[string]interface{}) int {
	switch v := body["code"].(type) {
	case float64:
		return int(v)
	default:
		if c, ok := body["Code"].(float64); ok {
			return int(c)
		}
	}
	return -999
}

func getOK(t *testing.T, rawURL string) string {
	t.Helper()
	resp, err := http.Get(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("%s -> %d %s", rawURL, resp.StatusCode, b)
	}
	return string(b)
}

func findJSONID(t *testing.T, raw string) string {
	t.Helper()
	var body struct {
		Rows []struct {
			ID uint
		}
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Rows) == 0 || body.Rows[0].ID == 0 {
		t.Fatalf("列表里没有 ID: %s", raw)
	}
	return strconv.FormatUint(uint64(body.Rows[0].ID), 10)
}

func findIDByName(t *testing.T, raw, name string) string {
	t.Helper()
	var body struct {
		Rows []struct {
			ID   uint
			Name string
		}
	}
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		t.Fatal(err)
	}
	for _, row := range body.Rows {
		if row.Name == name {
			return strconv.FormatUint(uint64(row.ID), 10)
		}
	}
	t.Fatalf("没有找到群 %s: %s", name, raw)
	return ""
}

func dialChat(t *testing.T, base string, user session) *websocket.Conn {
	t.Helper()
	wsURL := strings.Replace(base, "http", "ws", 1) + "/chat?userId=" + user.id + "&token=" + url.QueryEscape(user.token)
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		status := ""
		if resp != nil {
			status = resp.Status
			b, _ := io.ReadAll(resp.Body)
			status += " " + string(b)
		}
		t.Fatalf("连接聊天失败: %v %s", err, status)
	}
	return conn
}

func sendChat(t *testing.T, conn *websocket.Conn, from, to string, kind int, content string) {
	t.Helper()
	payload := map[string]interface{}{
		"userId":   jsonNumber(from),
		"TargetId": jsonNumber(to),
		"Type":     kind,
		"Media":    1,
		"Content":  content,
	}
	if err := conn.WriteJSON(payload); err != nil {
		t.Fatal(err)
	}
}

func jsonNumber(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func readChat(t *testing.T, conn *websocket.Conn) string {
	t.Helper()
	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func uploadTinyPNG(t *testing.T, rawURL string) string {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", "tiny.png")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = part.Write([]byte{0x89, 0x50, 0x4e, 0x47}); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, rawURL, &buf)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body struct {
		Code int
		Data string
		Msg  string
	}
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 {
		t.Fatalf("上传失败: %+v", body)
	}
	return body.Data
}
