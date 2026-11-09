package utils

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// Md5Encode 把任意字符串变成 32 位小写十六进制。
// 密码摘要用它。同样的明文每次结果一样，所以必须再加盐，否则相同密码的用户摘要会完全相同。
// MD5 很快，不适合保护真实密码，这里只为了看懂「盐 + 摘要」这条链路。
func Md5Encode(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

// MD5Encode 是同一段摘要的大写形式。
// 登录令牌用大写，密码摘要用小写，读库时能一眼分开这两类值。令牌仍然不是密码。
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// MakePassword 计算要写入 pass_word 的摘要。
// plainpwd 是用户刚输入的明文，salt 是这个用户自己的 6 位盐。两者直接拼接，中间没有分隔符。
func MakePassword(plainpwd, salt string) string {
	return Md5Encode(plainpwd + salt)
}

// ValidPassword 用库里的盐把刚输入的明文再算一遍。
// plainpwd 是本次输入，salt 和 password 都来自 user_basic。相等才算密码正确。
func ValidPassword(plainpwd, salt string, password string) bool {
	return Md5Encode(plainpwd+salt) == password
}
