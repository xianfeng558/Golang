package models

import (
	"ginchat/utils"

	"gorm.io/gorm"
)

// 好友申请的三种状态。同意之后才会写入 contact，页面上的直接加好友接口不受影响。
const (
	ApplyPending  = 0
	ApplyAccepted = 1
	ApplyRejected = 2
)

// FriendApply 是一条好友申请。FromID 发起，ToID 决定同意或拒绝。
type FriendApply struct {
	gorm.Model
	FromID uint   `gorm:"index;comment:申请人用户 ID"`
	ToID   uint   `gorm:"index;comment:被申请人用户 ID"`
	Status int    `gorm:"comment:0 待处理，1 已同意，2 已拒绝"`
	Remark string `gorm:"comment:申请附言"`
}

func (FriendApply) TableName() string { return "friend_apply" }

// ApplyFriend 按对方用户名发起申请。
// 已经是好友、已经有待处理申请、自己加自己，都不会插入。
func ApplyFriend(fromID uint, targetName, remark string) (int, string) {
	if targetName == "" {
		return -1, "好友用户名不能为空"
	}
	target := FindUserByName(targetName)
	if target.ID == 0 {
		return -1, "没有找到此用户"
	}
	if target.ID == fromID {
		return -1, "不能加自己"
	}
	var friend Contact
	utils.DB.Where("owner_id = ? and target_id = ? and type = 1", fromID, target.ID).Limit(1).Find(&friend)
	if friend.ID != 0 {
		return -1, "已经是好友"
	}
	var pending FriendApply
	utils.DB.Where("from_id = ? and to_id = ? and status = ?", fromID, target.ID, ApplyPending).Limit(1).Find(&pending)
	if pending.ID != 0 {
		return -1, "已经申请过了"
	}
	row := FriendApply{FromID: fromID, ToID: target.ID, Status: ApplyPending, Remark: remark}
	if err := utils.DB.Create(&row).Error; err != nil {
		return -1, "申请失败"
	}
	return 0, "申请已发出"
}

// ListApply 返回发给这个用户、还没处理的申请。
func ListApply(userID uint) []FriendApply {
	rows := make([]FriendApply, 0)
	utils.DB.Where("to_id = ? and status = ?", userID, ApplyPending).Order("id desc").Find(&rows)
	return rows
}

// ReplyApply 由被申请人处理。accept 为 true 时，在同一个事务里写成双向好友。
func ReplyApply(userID uint, applyID uint, accept bool) (int, string) {
	var row FriendApply
	utils.DB.Where("id = ?", applyID).Limit(1).Find(&row)
	if row.ID == 0 || row.ToID != userID {
		return -1, "申请不存在"
	}
	if row.Status != ApplyPending {
		return -1, "申请已经处理过"
	}
	if accept {
		var exist Contact
		utils.DB.Where("owner_id = ? and target_id = ? and type = 1", row.FromID, row.ToID).Limit(1).Find(&exist)
		if exist.ID != 0 {
			if err := utils.DB.Model(&FriendApply{}).Where("id = ?", row.ID).Update("status", ApplyAccepted).Error; err != nil {
				return -1, "处理申请失败"
			}
			return 0, "已同意"
		}
	}
	tx := utils.DB.Begin()
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	status := ApplyRejected
	msg := "已拒绝"
	if accept {
		status = ApplyAccepted
		msg = "已同意"
		if err := tx.Create(&Contact{OwnerId: row.FromID, TargetId: row.ToID, Type: 1}).Error; err != nil {
			return -1, "添加好友失败"
		}
		if err := tx.Create(&Contact{OwnerId: row.ToID, TargetId: row.FromID, Type: 1}).Error; err != nil {
			return -1, "添加好友失败"
		}
	}
	if err := tx.Model(&FriendApply{}).Where("id = ?", row.ID).Update("status", status).Error; err != nil {
		return -1, "处理申请失败"
	}
	if err := tx.Commit().Error; err != nil {
		return -1, "处理申请失败"
	}
	committed = true
	return 0, msg
}
