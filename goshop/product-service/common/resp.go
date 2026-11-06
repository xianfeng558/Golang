// Package common 是商品服务自己的一份 JSON 工具，和仓库根上的 common 模块不是同一个包。
// SwapToStruct 用 json 标签把数据库结构体抄到 proto 结构体，字段名对不上的会被丢掉。
package common

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/**
"{\"ID\":33,\"Name\":\"NeSugar小巢糖创意桌面无叶风扇usb台式风扇办公室喷雾风扇冷风机\",\"ProductType\":1,\"CategoryId\":52,\"StartingPrice\":92,\"TotalStock\":0,\"MainPicture\":\"https://msb-edu-dev.oss-cn-beijing.aliyuncs.com/mall-product/productO1CN01nsp1Wk1FIYRXXUXL8_!!4023510464-0-cib.jpg\",\"RemoteAreaPostage\":10,\"SingleBuyLimit\":0,\"IsEnable\":1,\"Remark\":\"NeSugar小巢糖创意桌面无叶风扇usb台式风扇办公室喷雾风扇冷风机\",\"CreateUser\":1,\"CreateTime\":\"2022-04-30T16:55:00+08:00\",\"Up"

**/
//通过json tag 进行结构体赋值
func SwapToStruct(req, target interface{}) (err error) {
	dataByte, err := json.Marshal(req)
	if err != nil {
		return
	}
	err = json.Unmarshal(dataByte, target)
	return
}

type H struct {
	Code                   string
	Message                string
	TraceId                string
	Data                   interface{}
	Rows                   interface{}
	Total                  interface{}
	SkyWalkingDynamicField string
}

func Resp(w http.ResponseWriter, code string, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code:    code,
		Data:    data,
		Message: message,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(ret)
}

func RespList(w http.ResponseWriter, code string, data interface{}, message string, rows interface{}, total interface{}, skyWalkingDynamicField string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	h := H{
		Code:                   code,
		Data:                   data,
		Message:                message,
		Rows:                   rows,
		Total:                  total,
		SkyWalkingDynamicField: skyWalkingDynamicField,
	}
	ret, err := json.Marshal(h)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(ret)
}

/**
200 OKLoginSuccessVO
201 Created
401 Unauthorized
403 Forbidden
404 Not Found
**/
func RespOK(w http.ResponseWriter, data interface{}, message string) {
	Resp(w, "SUCCESS", data, message)
}
func RespFail(w http.ResponseWriter, data interface{}, message string) {
	Resp(w, "TOKEN_FAIL", data, message)
}

//writer  data  message  row  total  field
func RespListOK(w http.ResponseWriter, data interface{}, message string, rows interface{}, total interface{}, skyWalkingDynamicField string) {
	RespList(w, "SUCCESS", data, message, rows, total, skyWalkingDynamicField)
}
func RespListFail(w http.ResponseWriter, data interface{}, message string, rows interface{}, total interface{}, skyWalkingDynamicField string) {
	RespList(w, "TOKEN_FAIL", data, message, rows, total, skyWalkingDynamicField)
}
