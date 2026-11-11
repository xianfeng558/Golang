package common

import (
	"math"
	"strconv"
	"strings"
)

// 这一组函数是课程里自带的“用户 ID ↔ 二进制字符串”转换。
// 登录时把用户 ID 变成一串 0/1 当作 Redis 的键，也当作后续请求头 Uuid。
// 它不是通用的进制库：ToInt 只按字符 '0' '1' 的 ASCII 做计算。
//
// 例子：用户 ID 是 1，ConverToBinary 得到 "1"，请求头 Uuid 就填 1。
// 用户 ID 是 2，得到 "10"，请求头 Uuid 要填 10。

// StringToArray 把字符串每个字符的 ASCII 码放进切片，再整段反转。
// 反转是为了让原来最右边的字符先被当成 2 的 0 次方。
func StringToArray(input string) []int {
	output := make([]int, 0, len(input))
	for _, v := range input {
		output = append(output, int(v))
	}
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}
	return output
}

// GetInput 把上面的 ASCII 码逐个送进通道。通道是 Go 里协程之间传递数据的管道。
func GetInput(input string) <-chan int {
	out := make(chan int)
	go func() {
		for _, b := range StringToArray(input) {
			out <- b
		}
		close(out)
	}()
	return out
}

// SQ 从通道读出 ASCII 码，减掉 '0' 的码值 48，再乘上 2 的递增次方。
// 这就是把二进制字符串还原成整数的过程，只是拆成了流水线。
func SQ(in <-chan int) <-chan int {
	out := make(chan int)
	var base, i float64 = 2, 0
	go func() {
		for n := range in {
			out <- (n - 48) * int(math.Pow(base, i))
			i++
		}
		close(out)
	}()
	return out
}

// ToInt 把二进制字符串还原成整数。参数名沿用课程里的拼写。
func ToInt(input string) int {
	sum := 0
	for o := range SQ(GetInput(input)) {
		sum += o
	}
	return sum
}

// ConverToBinary 把正整数变成二进制字符串，例如 2 变成 "10"。名字沿用课程原文。
func ConverToBinary(n int) string {
	if n == 0 {
		return "0"
	}
	res := ""
	for ; n > 0; n /= 2 {
		lsb := n % 2
		res = strconv.Itoa(lsb) + res
	}
	return res
}

// StockDeltaName 标记这一次改库存：Stock 字段是增减量，不是要写成的绝对库存。
// 负数表示扣减。商品服务用一条 SQL 做 stock = stock + Stock，并且拒绝把库存减成负数。
// 这个字符串只出现在 RPC 请求里，不会写入 product_sku.name。
const StockDeltaName = "__stock_delta__"

// SplitToInt32List 把 "12,35,66" 这种查询参数拆成整数切片。空片段和非法数字会被跳过。
func SplitToInt32List(str string, sep string) []int32 {
	tempStr := strings.Split(str, sep)
	int32List := make([]int32, 0, len(tempStr))
	for _, item := range tempStr {
		if item == "" {
			continue
		}
		val, err := strconv.ParseInt(item, 10, 32)
		if err != nil {
			continue
		}
		int32List = append(int32List, int32(val))
	}
	return int32List
}
