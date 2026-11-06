package common

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// H 是 HTTP 网关统一返回的 JSON 形状。
// Code 用 SUCCESS / TOKEN_FAIL 表示业务结果，HTTP 状态码在本项目里始终是 200。
type H struct {
	Code                   string      `json:"Code"`
	Message                string      `json:"Message"`
	TraceId                string      `json:"TraceId,omitempty"`
	Data                   interface{} `json:"Data"`
	Rows                   interface{} `json:"Rows,omitempty"`
	Total                  interface{} `json:"Total,omitempty"`
	SkyWalkingDynamicField string      `json:"SkyWalkingDynamicField,omitempty"`
}

// SwapToStruct 先把源对象编成 JSON，再解到目标对象。
// 字段对得上靠的是两边的 json 标签，而不是字段在内存里的顺序。
// 类型对不上时（例如 time.Time 对 string）对应字段会是零值。
func SwapToStruct(req, target interface{}) error {
	dataByte, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return json.Unmarshal(dataByte, target)
}

func writeJSON(w http.ResponseWriter, h H) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, _ = w.Write(ret)
}

// Resp 写一条普通 JSON。data 会放进 Data 字段。
func Resp(w http.ResponseWriter, code string, data interface{}, message string) {
	writeJSON(w, H{Code: code, Data: data, Message: message})
}

// RespList 比 Resp 多了分页字段 Rows、Total。商品列表接口用它。
func RespList(w http.ResponseWriter, code string, data interface{}, message string, rows interface{}, total interface{}, skyWalkingDynamicField string) {
	writeJSON(w, H{
		Code:                   code,
		Data:                   data,
		Message:                message,
		Rows:                   rows,
		Total:                  total,
		SkyWalkingDynamicField: skyWalkingDynamicField,
	})
}

// RespOK 表示业务成功。
func RespOK(w http.ResponseWriter, data interface{}, message string) {
	Resp(w, "SUCCESS", data, message)
}

// RespFail 表示业务失败。名字里的 TOKEN 来自课程，登录失败和库存不足都会用它。
func RespFail(w http.ResponseWriter, data interface{}, message string) {
	Resp(w, "TOKEN_FAIL", data, message)
}

// RespListOK 表示列表查询成功。
func RespListOK(w http.ResponseWriter, data interface{}, message string, rows interface{}, total interface{}, skyWalkingDynamicField string) {
	RespList(w, "SUCCESS", data, message, rows, total, skyWalkingDynamicField)
}

// RespListFail 表示列表查询失败。
func RespListFail(w http.ResponseWriter, data interface{}, message string, rows interface{}, total interface{}, skyWalkingDynamicField string) {
	RespList(w, "TOKEN_FAIL", data, message, rows, total, skyWalkingDynamicField)
}
