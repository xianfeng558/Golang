package models

import (
	"ginchat/utils"
)

// Contact 是一条关系，表名 contact。
// 好友会插两行：A 的名单里有 B，B 的名单里也有 A。群只插一行，成员指向群的数字 ID。
type Contact struct {
	Base
	OwnerId  uint   `gorm:"comment:这条关系属于谁"`
	TargetId uint   `gorm:"comment:好友时是对方用户 ID，加群时是群 ID"`
	Type     int    `gorm:"comment:1 好友，2 群成员"`
	Desc     string `gorm:"comment:备注"`
}

func (Contact) TableComment() string { return "已成立的好友或群关系" }

// TableName 固定成 contact。不写的话 GORM 会把它变成 contacts。
func (table *Contact) TableName() string {
	return "contact"
}

// SearchFriend 查出某人的好友，返回的是用户资料而不是关系行。
// 先按 owner_id 和 type=1 取出对方 ID，再按这些 ID 查 user_basic。
// 没有好友时返回空切片。若直接用空 ID 列表去做 IN 查询，不同版本的 SQL 表现不一致。
func SearchFriend(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contacts)
	if len(contacts) == 0 {
		return []UserBasic{}
	}
	ids := make([]uint, 0, len(contacts))
	for _, v := range contacts {
		ids = append(ids, v.TargetId)
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", ids).Find(&users)
	return users
}

// AddFriend 按对方用户名加好友。
// userId 是发起人，targetName 是对方登录名，不是对方的数字 ID。
// 返回值第一项 0 表示成功，-1 表示失败；第二项是页面要吐司出来的那句话。
// 自己加自己、重复加、找不到人，都在写库前返回。真正写入时两行放在同一个事务里，任何一行失败都会回滚。
func AddFriend(userId uint, targetName string) (int, string) {
	if targetName == "" {
		return -1, "好友用户名不能为空"
	}
	target := FindUserByName(targetName)
	if target.ID == 0 {
		return -1, "没有找到此用户"
	}
	if target.ID == userId {
		return -1, "不能加自己"
	}
	var exist Contact
	utils.DB.Where("owner_id = ? and target_id = ? and type = 1", userId, target.ID).Limit(1).Find(&exist)
	if exist.ID != 0 {
		return -1, "不能重复添加"
	}

	// committed 用来区分「正常提交」和「中途 return」。
	// 函数退出时如果还没提交，defer 会回滚，避免只插入了一边的好友关系。
	tx := utils.DB.Begin()
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&Contact{OwnerId: userId, TargetId: target.ID, Type: 1}).Error; err != nil {
		return -1, "添加好友失败"
	}
	if err := tx.Create(&Contact{OwnerId: target.ID, TargetId: userId, Type: 1}).Error; err != nil {
		return -1, "添加好友失败"
	}
	if err := tx.Commit().Error; err != nil {
		return -1, "添加好友失败"
	}
	committed = true
	return 0, "添加好友成功"
}

// JoinGroup 把用户加进一个群。
// comId 可以是群的数字 ID，也可以是群名称，页面上的输入框两种都接受。
// 查到群以后，关系里的 target_id 一律写群的数字 ID，不再把名称塞进这一列。
// 返回值和 AddFriend 一样：0 成功，-1 失败，后一项是说明。
func JoinGroup(userId uint, comId string) (int, string) {
	if userId == 0 || comId == "" {
		return -1, "群号不能为空"
	}
	community := Community{}
	utils.DB.Where("id = ? OR name = ?", comId, comId).Limit(1).Find(&community)
	if community.ID == 0 {
		return -1, "没有找到群"
	}
	var exist Contact
	utils.DB.Where("owner_id = ? AND target_id = ? AND type = 2", userId, community.ID).Limit(1).Find(&exist)
	if exist.ID != 0 {
		return -1, "已加过此群"
	}
	if err := utils.DB.Create(&Contact{OwnerId: userId, TargetId: community.ID, Type: 2}).Error; err != nil {
		return -1, "加群失败"
	}
	return 0, "加群成功"
}

// SearchUserByGroupId 返回群里所有成员的用户 ID，包含群主。
// 群发时拿这份名单，再由调用方跳过发送者本人，避免把消息又推回给自己。
func SearchUserByGroupId(communityId uint) []uint {
	contacts := make([]Contact, 0)
	utils.DB.Where("target_id = ? and type = 2", communityId).Find(&contacts)
	ids := make([]uint, 0, len(contacts))
	for _, v := range contacts {
		ids = append(ids, v.OwnerId)
	}
	return ids
}
