package models

import "gorm.io/gorm"

// GroupBasic 对应表 group_basic。
// 当前页面建群用的是 Community。这张表留在自动迁移里，避免以前已经建过的库缺表，业务路径不会读写它。
type GroupBasic struct {
	gorm.Model
	Name    string // 群名称
	OwnerId uint   // 群主用户 ID
	Icon    string // 群图标地址
	Type    int    // 预留的群分类，当前没有赋值约定
	Desc    string // 群说明
}

// TableName 固定成 group_basic，避免被收成 group_basics。
func (table *GroupBasic) TableName() string {
	return "group_basic"
}
