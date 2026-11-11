package model

import (
	"errors"
	"strings"
)

// UserAddress 是收货地址。一个用户可以有多条，其中最多一条 IsDefault 为 1。
// 下单时订单服务只保存地址 ID 的快照会更合理；当前订单表还没有地址列，这张表先把用户域补齐。
type UserAddress struct {
	ID        int32  `gorm:"primaryKey;comment:地址主键" json:"id"`
	UserID    int32  `gorm:"index;comment:用户 ID" json:"userId"`
	Receiver  string `gorm:"comment:收货人" json:"receiver"`
	Phone     string `gorm:"comment:收货手机" json:"phone"`
	Province  string `gorm:"comment:省" json:"province"`
	City      string `gorm:"comment:市" json:"city"`
	Detail    string `gorm:"comment:详细地址" json:"detail"`
	IsDefault int32  `gorm:"comment:1 默认地址，0 普通地址" json:"isDefault"`
	IsDeleted int32  `gorm:"comment:1 已删除" json:"isDeleted"`
}

func (UserAddress) TableName() string { return "user_address" }

// CheckAddress 检查一条地址能不能保存。不访问数据库。
func CheckAddress(addr UserAddress) error {
	if addr.UserID <= 0 {
		return errors.New("缺少用户")
	}
	if strings.TrimSpace(addr.Receiver) == "" {
		return errors.New("缺少收货人")
	}
	phone := strings.TrimSpace(addr.Phone)
	if len(phone) != 11 || phone[0] != '1' {
		return errors.New("手机号格式错误")
	}
	if strings.TrimSpace(addr.Detail) == "" {
		return errors.New("缺少详细地址")
	}
	return nil
}
