// Package repository 负责真正的数据库读写。上层 service 不写 SQL。
package repository

import (
	"errors"
	"fmt"

	"goproduct/domain/model"
	"goproduct/proto"
	"gorm.io/gorm"
)

// IProductRepository 是商品表的全部读写方法。handler 不直接依赖它，中间还隔着 service。
// *gorm.DB 是数据库连接。一个进程里通常共用这一份，GORM 内部有连接池。
type IProductRepository interface {
	Page(int32, int32) (int64, *[]model.Product, error)
	ShowProductDetail(int32) (*model.ProductDetail, error)
	ShowProductSku(int32) (*[]model.ProductSku, error)
	ShowDetailSku(int32) (*model.ProductSku, error)
	UpdateSku(req *proto.UpdateSkuReq) (isSuccess bool, err error)

	CountNum() int64
}

// 创建实例
func NewProductRepository(db *gorm.DB) IProductRepository {
	return &ProductRepository{mysqlDB: db}
}

// 数据DB
type ProductRepository struct {
	mysqlDB *gorm.DB
}

// Page 按页返回商品。length 是每页条数，pageIndex 从 1 开始。
// 这里新开查询链，不能把 Limit 写回 u.mysqlDB，否则后面的请求会一直带着上一次的分页条件。
func (u *ProductRepository) Page(length int32, pageIndex int32) (int64, *[]model.Product, error) {
	if length <= 0 || pageIndex <= 0 {
		return 0, &[]model.Product{}, errors.New("参数不匹配：length 和 pageIndex 都必须大于 0")
	}
	var count int64
	if err := u.mysqlDB.Model(&model.Product{}).Count(&count).Error; err != nil {
		return 0, nil, err
	}
	var arr []model.Product
	err := u.mysqlDB.Limit(int(length)).Offset((int(pageIndex) - 1) * int(length)).Find(&arr).Error
	if err != nil {
		fmt.Println("query product err :", err)
		return 0, nil, err
	}
	return count, &arr, nil
}

// ShowProductDetail 把商品主表、详情表、图片表拼成一条记录。
// GROUP BY 是为了通过 MySQL 8 的 ONLY_FULL_GROUP_BY 检查。
func (u *ProductRepository) ShowProductDetail(id int32) (product *model.ProductDetail, err error) {
	sql := "SELECT p.id, p.name, p.product_type, p.category_id, p.starting_price, p.main_picture, " +
		"IFNULL(pd.detail,'') AS detail, IFNULL(GROUP_CONCAT(pp.picture SEPARATOR ','), '') AS picture_list " +
		"FROM product p " +
		"LEFT JOIN product_detail pd ON p.id = pd.product_id " +
		"LEFT JOIN product_picture pp ON p.id = pp.product_id " +
		"WHERE p.id = ? " +
		"GROUP BY p.id, p.name, p.product_type, p.category_id, p.starting_price, p.main_picture, pd.detail"
	var productDetails []model.ProductDetail
	if err = u.mysqlDB.Raw(sql, id).Scan(&productDetails).Error; err != nil {
		return nil, err
	}
	if len(productDetails) == 0 {
		return nil, errors.New("商品不存在")
	}
	fmt.Println("repository ShowProductDetail   >>>> ", productDetails[0].ID)
	return &productDetails[0], nil
}
func (u *ProductRepository) CountNum() int64 {
	var count int64
	u.mysqlDB.Model(&model.Product{}).Offset(-1).Limit(-1).Count(&count)
	return count
}

/*
SkuId               int32 `gorm:"column:id" json:"skuId"`
Name                string
AttributeSymbolList string  `gorm:"column:attribute_symbolList" json:"attributeSymbolList"`
SellPrice           float32 `gorm:"column:sell_price" json:"sellPrice"`
Stock               int32   `gorm:"default:1"`
*/
func (u *ProductRepository) ShowProductSku(id int32) (product *[]model.ProductSku, err error) {
	sql := "select id ,name ,attribute_symbol_list,sell_price,stock  from product_sku where product_id= ?"
	var productSku []model.ProductSku

	u.mysqlDB.Raw(sql, id).Scan(&productSku)
	fmt.Println("repository ShowProductSku   >>>> ", productSku)
	return &productSku, nil
}

func (u *ProductRepository) ShowDetailSku(id int32) (obj *model.ProductSku, err error) {
	var productSku = &model.ProductSku{}
	u.mysqlDB.Where("id = ?", id).Find(&productSku)
	fmt.Println("repository ShowDetailSku   >>>> ", productSku)
	return productSku, nil
}

func (u *ProductRepository) UpdateSku(req *proto.UpdateSkuReq) (isSuccess bool, err error) {
	sku := req.GetProductSku()
	isSuccess = true
	//u.mysqlDB.Updates(sku)
	tb := u.mysqlDB.Debug().Model(&model.ProductSku{}).Where("id=?", sku.SkuId).Update("stock", sku.Stock)
	if tb.Error != nil {
		isSuccess = false
	}
	return isSuccess, tb.Error
}
