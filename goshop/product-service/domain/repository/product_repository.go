// Package repository 负责真正的数据库读写。上层 service 不写 SQL。
package repository

import (
	"errors"

	common "git.mashibing.com/msb_47094/shopping-comm"
	"goproduct/domain/model"
	"goproduct/proto"
	"gorm.io/gorm"
)

// IProductRepository 是商品表的全部读写方法。
// *gorm.DB 是共用的数据库连接，GORM 内部有连接池。
type IProductRepository interface {
	Page(int32, int32) (int64, *[]model.Product, error)
	ShowProductDetail(int32) (*model.ProductDetail, error)
	ShowProductSku(int32) (*[]model.ProductSku, error)
	ShowDetailSku(int32) (*model.ProductSku, error)
	UpdateSku(req *proto.UpdateSkuReq) (isSuccess bool, err error)
}

// NewProductRepository 把连接放进结构体，并按接口类型返回。
func NewProductRepository(db *gorm.DB) IProductRepository {
	return &ProductRepository{mysqlDB: db}
}

// ProductRepository 保存这一份 MySQL 连接。小写字段别的包看不见。
type ProductRepository struct {
	mysqlDB *gorm.DB
}

// Page 按页返回商品。length 是每页条数，pageIndex 从 1 开始。
// Limit 写在本次查询链上，不能写回 u.mysqlDB，否则后面的请求会带着上一次的分页条件。
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
		return 0, nil, err
	}
	return count, &arr, nil
}

// ShowProductDetail 把商品主表、详情表、图片表拼成一条记录。
// GROUP BY 是为了通过 MySQL 8 的 ONLY_FULL_GROUP_BY 检查。
func (u *ProductRepository) ShowProductDetail(id int32) (*model.ProductDetail, error) {
	sql := "SELECT p.id, p.name, p.product_type, p.category_id, p.starting_price, p.main_picture, " +
		"IFNULL(pd.detail,'') AS detail, IFNULL(GROUP_CONCAT(pp.picture SEPARATOR ','), '') AS picture_list " +
		"FROM product p " +
		"LEFT JOIN product_detail pd ON p.id = pd.product_id " +
		"LEFT JOIN product_picture pp ON p.id = pp.product_id " +
		"WHERE p.id = ? " +
		"GROUP BY p.id, p.name, p.product_type, p.category_id, p.starting_price, p.main_picture, pd.detail"
	var productDetails []model.ProductDetail
	if err := u.mysqlDB.Raw(sql, id).Scan(&productDetails).Error; err != nil {
		return nil, err
	}
	if len(productDetails) == 0 {
		return nil, errors.New("商品不存在")
	}
	return &productDetails[0], nil
}

// ShowProductSku 查出某个商品下的全部规格。没有规格时返回空切片，不是错误。
func (u *ProductRepository) ShowProductSku(id int32) (*[]model.ProductSku, error) {
	sql := "select id ,name ,attribute_symbol_list,sell_price,stock  from product_sku where product_id= ?"
	var productSku []model.ProductSku
	if err := u.mysqlDB.Raw(sql, id).Scan(&productSku).Error; err != nil {
		return nil, err
	}
	return &productSku, nil
}

// ShowDetailSku 按规格主键查一条。找不到时返回错误，加购流程会提示「查询规格失败」。
func (u *ProductRepository) ShowDetailSku(id int32) (*model.ProductSku, error) {
	productSku := &model.ProductSku{}
	if err := u.mysqlDB.Where("id = ?", id).Find(productSku).Error; err != nil {
		return nil, err
	}
	if productSku.SkuId == 0 {
		return nil, errors.New("规格不存在")
	}
	return productSku, nil
}

// UpdateSku 改某一条规格的库存。
//
// 普通请求把 Stock 当成目标库存，直接覆盖。两个请求同时读到 10、都写回 9 时，会少扣一次。
// 加购不走这条覆盖路径。请求的 Name 等于 common.StockDeltaName 时，Stock 是增减量：
// 一条 UPDATE 同时完成「够不够」和「减去多少」。后到的请求如果会减成负数，影响行数是 0，库存保持原样。
// 数据库对这一行加锁，两个加购不会读到同一份旧库存再各自写回。
func (u *ProductRepository) UpdateSku(req *proto.UpdateSkuReq) (bool, error) {
	sku := req.GetProductSku()
	if sku == nil || sku.SkuId == 0 {
		return false, errors.New("规格不存在")
	}
	if sku.Name == common.StockDeltaName {
		res := u.mysqlDB.Exec(
			"UPDATE product_sku SET stock = stock + ? WHERE id = ? AND stock + ? >= 0",
			sku.Stock, sku.SkuId, sku.Stock,
		)
		if res.Error != nil {
			return false, res.Error
		}
		if res.RowsAffected == 0 {
			return false, errors.New("库存不足或规格不存在")
		}
		return true, nil
	}
	tb := u.mysqlDB.Model(&model.ProductSku{}).Where("id=?", sku.SkuId).Update("stock", sku.Stock)
	return tb.Error == nil, tb.Error
}
