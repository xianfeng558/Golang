// Package handler 是 RPC 的入口。proto 里声明的 Login、GetUserToken 最终调到这里。
// ctx 用来传递超时和链路信息；req 是客户端发来的参数；resp 由本方法填好后返回。
package handler

import (
	"context"
	"fmt"
	common "git.mashibing.com/msb_47094/shopping-comm"
	"gouser/domain/model"
	"gouser/domain/service"
	"gouser/proto"
	"strconv"
	"time"
)

// User 是登录 RPC 的实现。同一个结构体可以同时实现 Login 和 GetUserToken 两个服务。
type User struct {
	UserDataService service.IUserDataService
}

// Login 先按手机号查找或注册，再把用户信息填进 loginResp，最后把令牌写入 Redis，有效期 1 小时。
// GetPhone 和直接读 .Phone 等价，Get 方法是 protobuf 生成的，空指针时返回零值，更安全。
// strconv.Itoa 把整数用户 ID 变成字符串，当作 Redis 的键。
func (u *User) Login(ctx context.Context, loginRequest *proto.LoginRequest, loginResp *proto.LoginResp) error {
	userInfo, err := u.UserDataService.Login(loginRequest.ClientId, loginRequest.GetPhone(), loginRequest.SystemId, loginRequest.VerificationCode)
	if err != nil {
		return err
	}
	UserForResp(userInfo, loginResp)
	u.UserDataService.SetUserToken(strconv.Itoa(int(userInfo.ID)), []byte(loginResp.Token), time.Duration(1)*time.Hour)
	return nil
}

// UserForResp 把数据库里的 User 抄到响应里，并用当前时间的 MD5 当作登录令牌。
// 这个令牌不是密码。正式项目应使用随机串，并设置过期时间由 Redis 的 TTL 控制。
func UserForResp(userModel *model.User, resp *proto.LoginResp) *proto.LoginResp {
	timeStr := fmt.Sprintf("%d", time.Now().Unix())
	resp.Token = common.Md5Encode(timeStr) //"123456"
	resp.User = &proto.User{}
	resp.User.Id = userModel.ID
	resp.User.Avatar = userModel.Avatar
	resp.User.ClientId = userModel.ClientId
	resp.User.EmployeeId = 1 //userModel.EmployeeId
	resp.User.Nickname = userModel.Nickname
	resp.User.SessionId = resp.Token
	resp.User.Phone = userModel.Phone
	//token 过期时间
	tp, _ := time.ParseDuration("1h")
	tokenExpireTime := time.Now().Add(tp)
	expiretimeStr := tokenExpireTime.Format("2006-01-02 15:04:05")
	resp.User.TokenExpireTime = expiretimeStr
	resp.User.UnionId = userModel.UnionId
	return resp
}

// GetUserToken 给购物车和订单用：拿请求里的 Uuid 去 Redis 查令牌。
// 查到了就认为已登录，并把过期时间再延长 1 小时。查不到就把 IsLogin 设为 false。
func (u *User) GetUserToken(ctx context.Context, req *proto.TokenReq, resp *proto.TokenResp) error {
	res := u.UserDataService.GetUserToken(req.GetUuid())
	if res != "" {
		resp.IsLogin = true
		resp.Token = res
		// 续命
		uuid := common.ToInt(req.Uuid)
		u.UserDataService.SetUserToken(strconv.Itoa(uuid), []byte(res), time.Duration(1)*time.Hour)
	} else {
		resp.IsLogin = false
		resp.Token = ""
	}
	return nil
}
