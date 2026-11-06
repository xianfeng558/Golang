// Package repository 读写 shopping_cart。GetOrderTotal 会 JOIN 商品库里的 product_sku。
package repository

import (
	"fmt"
	"shoppingCart/domain/model"
	"shoppingCart/proto"
	"time"

	"gorm.io/gorm"
)

/**
 int32 clientId = 1;
    string phone = 2;
    int32 systemId = 3;
    string verificationCode = 4;
**/
//接口
type ICartRepository interface {
	AddCart(*proto.AddCartReq) (*model.ShoppingCart, error)
	UpdateCart(req *proto.UpdateCartReq) (*model.ShoppingCart, error)
	GetOrderTotal(int32List []int32) (obj float32, err error)
	FindCart(*proto.FindCartReq) (obj *model.ShoppingCart, err error)
}

// 创建实例
func NewCartRepository(db *gorm.DB) ICartRepository {
	return &CartRepository{mysqlDB: db}
}

// 数据DB
type CartRepository struct {
	mysqlDB *gorm.DB
}

func (u *CartRepository) AddCart(req *proto.AddCartReq) (obj *model.ShoppingCart, err error) {
	cart := model.ShoppingCart{
		Number:             req.Number,
		ProductId:          req.ProductId,
		ProductSkuId:       req.ProductSkuId,
		ProductName:        req.ProductName,
		ProductMainPicture: req.ProductMainPicture,
		UserId:             req.UserId,
		CreateUser:         req.CreateUser,
	}
	cart.CreateTime = time.Now() //
	tb := u.mysqlDB.Create(&cart)
	//err = errors.New("400", "测试异常", 400)
	fmt.Println("repository AddCart   >>>> ", cart)
	return &cart, tb.Error //err
}

func (u *CartRepository) UpdateCart(req *proto.UpdateCartReq) (obj *model.ShoppingCart, err error) {
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
	}
	cart.UpdateTime = time.Now() //  &cart
	tb := u.mysqlDB.Model(&model.ShoppingCart{}).
		Where("id = ?", cart.ID).
		Updates(&cart)
	//Update("is_deleted", 1)
	//err = errors.New("400", "测试异常", 400)
	fmt.Println("repository UpdateCart   >>>> ", cart)
	return &cart, tb.Error //err
}

// 统计订单价格
func (u *CartRepository) GetOrderTotal(int32List []int32) (obj float32, err error) {
	// number、sell_price 是 GORM 默认生成的列名。购物车和 SKU 必须在同一个库，JOIN 才找得到表。
	sql := "select sum(c.number * s.sell_price) from `shopping_cart` c " +
		"LEFT JOIN `product_sku` s on c.product_sku_id=s.id " +
		"where c.id in ?"
	var totalPrice float32
	tb := u.mysqlDB.Raw(sql, int32List).Scan(&totalPrice)
	//Update("order_status", trade.OrderStatus)
	//err = errors.New("400", "测试异常", 400)
	fmt.Println("GetOrderTotal   >>>> ", totalPrice)
	return totalPrice, tb.Error //err
}

// 查询购物车
func (u *CartRepository) FindCart(req *proto.FindCartReq) (obj *model.ShoppingCart, err error) {
	id := req.Id
	cart := &model.ShoppingCart{}

	tb := u.mysqlDB.Where("id = ?", id).Find(cart)
	fmt.Println("FindCart     >>>> ", cart)
	return cart, tb.Error //err
}
