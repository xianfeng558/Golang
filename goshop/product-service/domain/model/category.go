package model

// Category 是商品分类。商品表里的 category_id 指向这里。
// 示例商品使用分类 31，种子数据会保证这一行存在。
type Category struct {
	ID       int32  `gorm:"primaryKey;comment:分类主键" json:"id"`
	ParentID int32  `gorm:"index;comment:父分类，0 表示顶级" json:"parentId"`
	Name     string `gorm:"comment:分类名称" json:"name"`
	Sort     int32  `gorm:"comment:排序，小的靠前" json:"sort"`
	IsEnable bool   `gorm:"comment:是否在前台展示" json:"isEnable"`
}

func (Category) TableName() string { return "product_category" }
