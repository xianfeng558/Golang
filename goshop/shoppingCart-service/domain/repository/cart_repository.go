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

// AddCart 插入一行。req.Id 大于 0 时使用调用方事先生成的主键。
// 分布式事务的正向和补偿是两次独立 HTTP，补偿请求里必须已经带着这一行的 ID。
// 如果等插入成功后再把自增 ID 存进进程变量，两个用户同时下单会互相覆盖。
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
	if req.Id > 0 {
		cart.ID = req.Id
	}
	tb := u.mysqlDB.Create(&cart)
	return &cart, tb.Error
}

// UpdateCart 按主键更新购物车。
// GORM 用结构体 Updates 时会跳过零值，bool 的 false 也算零值。
// 订单补偿要把 is_deleted 从 true 改回 false，不指定列的话这一列写不进去，购物车会一直保持删除。
// 只把请求里真正带了的列放进 Select，避免补偿请求里的空字符串把商品名清掉。
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
	db := u.mysqlDB.Model(&model.ShoppingCart{}).Where("id = ?", cart.ID)
	if !req.IsDeleted {
		cols := []string{"IsDeleted", "UpdateTime"}
		if req.Number != 0 {
			cols = append(cols, "Number")
		}
		if req.ProductId != 0 {
			cols = append(cols, "ProductId")
		}
		if req.ProductSkuId != 0 {
			cols = append(cols, "ProductSkuId")
		}
		if req.ProductName != "" {
			cols = append(cols, "ProductName")
		}
		if req.ProductMainPicture != "" {
			cols = append(cols, "ProductMainPicture")
		}
		if req.UserId != 0 {
			cols = append(cols, "UserId")
		}
		if req.UpdateUser != 0 {
			cols = append(cols, "UpdateUser")
		}
		db = db.Select(cols)
	}
	tb := db.Updates(&cart)
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
