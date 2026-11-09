package biz

import (
	"context"

	v1 "valuation/api/helloworld/v1"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

var (
	// ErrUserNotFound 是脚手架的「找不到用户」。代驾接口不用这个错误。
	ErrUserNotFound = errors.NotFound(v1.ErrorReason_USER_NOT_FOUND.String(), "user not found")
)

// Greeter 是 Kratos 脚手架留下的问候示例，不参与代驾业务。
// wire 仍然引用它。删掉 NewGreeterUsecase 会让生成的 wire_gen.go 编译失败。
type Greeter struct {
	Hello string
}

// GreeterRepo 是脚手架的存储接口。代驾业务不调用这五个方法。
type GreeterRepo interface {
	Save(context.Context, *Greeter) (*Greeter, error)
	Update(context.Context, *Greeter) (*Greeter, error)
	FindByID(context.Context, int64) (*Greeter, error)
	ListByHello(context.Context, string) ([]*Greeter, error)
	ListAll(context.Context) ([]*Greeter, error)
}

// GreeterUsecase 是脚手架的用例，只把问候写进日志再交给 repo。
type GreeterUsecase struct {
	repo GreeterRepo
	log  *log.Helper
}

// NewGreeterUsecase 供 wire 组装。不要从 ProviderSet 里删掉它。
func NewGreeterUsecase(repo GreeterRepo, logger log.Logger) *GreeterUsecase {
	return &GreeterUsecase{repo: repo, log: log.NewHelper(logger)}
}

// CreateGreeter 把问候交给 repo 保存，并返回保存后的对象。代驾流程不会调用它。
func (uc *GreeterUsecase) CreateGreeter(ctx context.Context, g *Greeter) (*Greeter, error) {
	uc.log.WithContext(ctx).Infof("CreateGreeter: %v", g.Hello)
	return uc.repo.Save(ctx, g)
}
