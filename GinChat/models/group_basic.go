package models

// GroupBasic 对应表 group_basic。
// 当前页面建群用的是 Community。这张表留在自动迁移里，避免以前已经建过的库缺表，业务路径不会读写它。
type GroupBasic struct {
	Base
	Name    string `gorm:"comment:群名称"`
	OwnerId uint   `gorm:"comment:群主用户 ID"`
	Icon    string `gorm:"comment:群图标地址"`
	Type    int    `gorm:"comment:预留的群分类"`
	Desc    string `gorm:"comment:群说明"`
}

func (GroupBasic) TableComment() string { return "旧群表，当前建群不写它" }

// TableName 固定成 group_basic，避免被收成 group_basics。
func (table *GroupBasic) TableName() string {
	return "group_basic"
}
