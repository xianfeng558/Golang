// Package model 是购物车表 shopping_cart。它和商品 SKU 放在同一个 product_center 库。
package model

import "time"

type ShoppingCart struct {
	//gorm.Model
	ID                 int32     `json:"id"`
	UserId             int32     `gorm:"default:1" json:"userId"`
	ProductId          int32     `gorm:"product_id" json:"productId"`
	ProductSkuId       int32     `gorm:"product_sku_id" json:"productSkuId"`
	ProductName        string    `json:"productName"`
	ProductMainPicture string    `gorm:"product_main_picture" json:"productMainPicture"`
	Number             int32     `gorm:"default:1" json:"shoppingCartNumber"`
	CreateUser         int32     `gorm:"default:1" json:"createUser"`
	CreateTime         time.Time `json:"createTime"`
	UpdateUser         int32     `json:"updateUser"`
	UpdateTime         time.Time `json:"updateTime"`
	IsDeleted          bool      `json:"isDeleted"`
}

func (table *ShoppingCart) TableName() string {
	return "shopping_cart"
}
