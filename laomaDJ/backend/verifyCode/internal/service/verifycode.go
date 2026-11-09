// Package service 按长度和类型生成验证码。
// 本服务不写 Redis，也不校验手机号。乘客、司机拿到码之后各自保存，并在登录时自己比对。
package service

import (
	"context"
	"math/rand"

	pb "verifyCode/api/verifyCode"
)

// VerifyCodeService 生成验证码。它不存码，存到 Redis 是乘客和司机服务的事。
// 乘客用 Redis 1 号库、键 CVC:；司机用 2 号库、键 DVC:。两边互不影响。
type VerifyCodeService struct {
	pb.UnimplementedVerifyCodeServer
}

// NewVerifyCodeService 供 wire 组装。
func NewVerifyCodeService() *VerifyCodeService {
	return &VerifyCodeService{}
}

// GetVerifyCode 按长度和类型生成一串验证码。
// Length 小于等于 0 时按 6 位。大于 32 时截成 32，避免一次请求生成过长的随机串。
// Type 为 1（DIGIT）时是纯数字，乘客和司机目前都请求这一类。0 是 DEFAULT，字符集和数字相同。
func (s *VerifyCodeService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeRequest) (*pb.GetVerifyCodeReply, error) {
	length := int(req.Length)
	if length <= 0 {
		length = 6
	}
	if length > 32 {
		length = 32
	}
	return &pb.GetVerifyCodeReply{Code: RandCode(length, req.Type)}, nil
}

// RandCode 按类型从对应字符集里取指定长度的随机串。
// DIGIT 和 DEFAULT 用 0-9。LETTER 用小写字母。MIXED 用数字加小写字母。
// 未定义的类型返回空串，不猜测成数字，避免调用方把错误类型当成合法验证码。
func RandCode(length int, kind pb.TYPE) string {
	switch kind {
	case pb.TYPE_DEFAULT, pb.TYPE_DIGIT:
		return randCode("0123456789", length, 4)
	case pb.TYPE_LETTER:
		return randCode("abcdefghijklmnopqrstuvwxyz", length, 5)
	case pb.TYPE_MIXED:
		return randCode("0123456789abcdefghijklmnopqrstuvwxyz", length, 6)
	default:
		return ""
	}
}

// randCode 一次取出 63 位随机数，按字符集需要的位数切成多段来用。
// idxBits 必须盖住字符集下标：10 个数字用 4 位（0–15），26 个字母用 5 位，36 个混合字符用 6 位。
// 切出来的数如果大于等于字符集长度，就丢掉再取下一段。
// 这样每个字符被选中的机会相同。如果直接用下标对长度取模，余数较小的字符会被选得更多。
// math/rand 从 Go 1.20 起自动播种，这里不再调用 rand.Seed。
func randCode(chars string, length, idxBits int) string {
	idxMask := int64(1<<idxBits - 1)
	idxMax := 63 / idxBits
	result := make([]byte, length)
	for i, cache, remain := 0, rand.Int63(), idxMax; i < length; {
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		if idx := int(cache & idxMask); idx < len(chars) {
			result[i] = chars[idx]
			i++
		}
		cache >>= idxBits
		remain--
	}
	return string(result)
}
