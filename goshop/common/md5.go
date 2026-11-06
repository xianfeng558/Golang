package common

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// Md5Encode 计算字符串的 MD5，结果是 32 位小写十六进制。
// 本项目用它生成登录 token。MD5 已经不适合保护密码，正式项目应使用 bcrypt 或类似算法。
func Md5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// MD5Encode 是大写形式的 MD5。当前业务没有调用。
func MD5Encode(data string) string {
	return strings.ToUpper(Md5Encode(data))
}

// MakePassword 把明文和盐拼在一起再做 MD5。盐用来让相同密码得到不同哈希。
func MakePassword(plainpwd, salt string) string {
	return Md5Encode(plainpwd + salt)
}

// ValidPassword 用同样的规则重算哈希，和库存的哈希比较。相等就认为密码正确。
func ValidPassword(plainpwd, salt string, password string) bool {
	return Md5Encode(plainpwd+salt) == password
}
