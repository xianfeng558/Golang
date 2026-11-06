// Package repository 读写 shopping_cart。GetOrderTotal 会 JOIN 同一个库里的 product_sku。
package repository

import (
	"shoppingCart/domain/model"
	"shoppingCart/proto"
	"time"

	"gorm.io/gorm"
)

// ICartRepository 是购物车表的增、改、查和算价。
type ICartRepository interface {
	AddCart(*proto.AddCartReq) (*model.ShoppingCart, error)
	UpdateCart(req *proto.UpdateCartReq) (*model.ShoppingCart, error)
	GetOrderTotal(int32List []int32) (obj float32, err error)
	FindCart(*proto.FindCartReq) (obj *model.ShoppingCart, err error)
}

func NewCartRepository(db *gorm.DB) ICartRepository {
	return &CartRepository{mysqlDB: db}
}

type CartRepository struct {
	mysqlDB *gorm.DB
}

func (u *CartRepository) AddCart(req *proto.AddCartReq) (*model.ShoppingCart, error) {
	cart := model.ShoppingCart{
		Number:             req.Number,
		ProductId:          req.ProductId,
		ProductSkuId:       req.ProductSkuId,
		ProductName:        req.ProductName,
		ProductMainPicture: req.ProductMainPicture,
		UserId:             req.UserId,
		CreateUser:         req.CreateUser,
		CreateTime:         time.Now(),
	}
	tb := u.mysqlDB.Create(&cart)
	return &cart, tb.Error
}

func (u *CartRepository) UpdateCart(req *proto.UpdateCartReq) (*model.ShoppingCart, error) {
	cart := model.ShoppingCart{
		Number:             req.Number,
		ProductId:          req.ProductId,
		ProductSkuId:       req.ProductSkuId,
		ProductName:        req.ProductName,
		ProductMainPicture: req.ProductMainPicture,
		UserId:             req.UserId,
		ID:                 req.Id,
		IsDeleted:          req.IsDeleted,
		UpdateUser:         req.UpdateUser,
		UpdateTime:         time.Now(),
	}
	tb := u.mysqlDB.Model(&model.ShoppingCart{}).Where("id = ?", cart.ID).Updates(&cart)
	return &cart, tb.Error
}

// GetOrderTotal 用数量乘规格售价再求和。购物车和 SKU 必须在同一个库，JOIN 才找得到表。
func (u *CartRepository) GetOrderTotal(int32List []int32) (float32, error) {
	sql := "select sum(c.number * s.sell_price) from `shopping_cart` c " +
		"LEFT JOIN `product_sku` s on c.product_sku_id=s.id " +
		"where c.id in ?"
	var totalPrice float32
	tb := u.mysqlDB.Raw(sql, int32List).Scan(&totalPrice)
	return totalPrice, tb.Error
}

func (u *CartRepository) FindCart(req *proto.FindCartReq) (*model.ShoppingCart, error) {
	cart := &model.ShoppingCart{}
	tb := u.mysqlDB.Where("id = ?", req.Id).Find(cart)
	return cart, tb.Error
}
