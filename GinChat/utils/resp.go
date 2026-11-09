package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// H 是好友、群、上传、历史消息用的 JSON。
// 字段名是大写开头，因为页面里的判断写的是 res.Code、res.Rows、res.Data。
// 登录和注册不走这个结构，它们用小写的 code、message、data，对应页面上的 res.code。
type H struct {
	Code  int         `json:"Code"`  // 0 成功，-1 失败。页面判断的是这个字段，不看 HTTP 状态码
	Msg   string      `json:"Msg"`   // 给人看的一句话，例如「添加好友成功」
	Data  interface{} `json:"Data"`  // 单条结果。查用户时是用户对象，加好友成功时是数字 0
	Rows  interface{} `json:"Rows"`  // 列表。好友、群、历史消息都放这里
	Total interface{} `json:"Total"` // 好友列表里是条数，群列表里是「查询成功」，历史消息里是 Rows 的长度
}

// writeJSON 固定回 200。业务失败也走这里，调用方把失败写进 Code。
func writeJSON(w http.ResponseWriter, h H) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	body, err := json.Marshal(h)
	if err != nil {
		log.Println("json marshal", err)
		return
	}
	if _, err = w.Write(body); err != nil {
		log.Println("write json", err)
	}
}

// Resp 写单条结果。code 为 0 表示成功，-1 表示失败。HTTP 状态仍是 200，页面只看 Code。
func Resp(w http.ResponseWriter, code int, data interface{}, msg string) {
	writeJSON(w, H{Code: code, Data: data, Msg: msg})
}

// RespList 写列表。Rows 是数组，Total 在好友列表里是数量，在群列表里是一句说明。
func RespList(w http.ResponseWriter, code int, data interface{}, total interface{}) {
	writeJSON(w, H{Code: code, Rows: data, Total: total})
}

// RespFail 失败时 Code 为 -1，Data 为空。
func RespFail(w http.ResponseWriter, msg string) {
	Resp(w, -1, nil, msg)
}

// RespOK 成功时 Code 为 0。
func RespOK(w http.ResponseWriter, data interface{}, msg string) {
	Resp(w, 0, data, msg)
}

// RespOKList 成功的列表。data 放进 Rows，total 放进 Total。
func RespOKList(w http.ResponseWriter, data interface{}, total interface{}) {
	RespList(w, 0, data, total)
}
