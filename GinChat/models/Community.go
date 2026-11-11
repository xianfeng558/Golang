package models

import (
	"ginchat/utils"
)

// Community 是一个群，对应表 communities。
// 页面上的「创建群聊」写的是这张表。group_basic 是另一张留着的旧表，当前建群不写它。
type Community struct {
	Base
	Name    string `gorm:"comment:群名称"`
	OwnerId uint   `gorm:"comment:群主用户 ID"`
	Img     string `gorm:"comment:群头像地址"`
	Desc    string `gorm:"comment:群简介"`
}

func (Community) TableComment() string { return "群" }

// TableName 固定成 communities，和 GORM 默认的复数形式一致，写出来是为了读代码时不用猜。
func (table *Community) TableName() string {
	return "communities"
}

// CreateCommunity 建群，并把群主写成一条 type=2 的关系。
// 群和关系在同一个事务里。群插入成功后 GORM 会把自增 ID 填回 community.ID，关系的 target_id 用的就是它。
// 群主因此会出现在自己的群列表里，群发时也会出现在成员名单里。
// 返回 0 和「建群成功」，或 -1 和失败原因。
func CreateCommunity(community Community) (int, string) {
	if community.Name == "" {
		return -1, "群名称不能为空"
	}
	if community.OwnerId == 0 {
		return -1, "请先登录"
	}
	tx := utils.DB.Begin()
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()
	if err := tx.Create(&community).Error; err != nil {
		return -1, "建群失败"
	}
	contact := Contact{OwnerId: community.OwnerId, TargetId: community.ID, Type: 2}
	if err := tx.Create(&contact).Error; err != nil {
		return -1, "添加群关系失败"
	}
	if err := tx.Commit().Error; err != nil {
		return -1, "建群失败"
	}
	committed = true
	return 0, "建群成功"
}

// LoadCommunity 返回这个人加入过的群，包括他自己创建的。
// 先从 contact 里取 type=2 的 target_id，再查 communities。
// 一个群都没有时返回空切片和「查询成功」。这不是失败，页面会把空列表显示成没有群。
func LoadCommunity(ownerId uint) ([]Community, string) {
	contacts := make([]Contact, 0)
	utils.DB.Where("owner_id = ? and type = 2", ownerId).Find(&contacts)
	if len(contacts) == 0 {
		return []Community{}, "查询成功"
	}
	ids := make([]uint, 0, len(contacts))
	for _, v := range contacts {
		ids = append(ids, v.TargetId)
	}
	data := make([]Community, 0)
	utils.DB.Where("id in ?", ids).Find(&data)
	return data, "查询成功"
}
