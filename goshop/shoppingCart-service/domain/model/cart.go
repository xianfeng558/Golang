// Package model 是购物车表 shopping_cart。它和商品 SKU 放在同一个 product_center 库。
package model

import "time"

type ShoppingCart struct {
	ID                 int32     `gorm:"primaryKey;comment:购物车主键" json:"id"`
	UserId             int32     `gorm:"default:1;comment:用户 ID" json:"userId"`
	ProductId          int32     `gorm:"comment:商品 ID" json:"productId"`
	ProductSkuId       int32     `gorm:"comment:规格 ID" json:"productSkuId"`
	ProductName        string    `gorm:"comment:加购时的商品名" json:"productName"`
	ProductMainPicture string    `gorm:"comment:加购时的封面" json:"productMainPicture"`
	Number             int32     `gorm:"default:1;comment:数量" json:"shoppingCartNumber"`
	CreateUser         int32     `gorm:"default:1;comment:创建人" json:"createUser"`
	CreateTime         time.Time `gorm:"comment:创建时间" json:"createTime"`
	UpdateUser         int32     `gorm:"comment:更新人" json:"updateUser"`
	UpdateTime         time.Time `gorm:"comment:更新时间" json:"updateTime"`
	IsDeleted          bool      `gorm:"comment:软删除标记" json:"isDeleted"`
}

func (ShoppingCart) TableComment() string { return "购物车" }

func (table *ShoppingCart) TableName() string {
	return "shopping_cart"
}
