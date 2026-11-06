// Package service 把商品查询和改库存转交给 repository。
// 接口 IProductDataService 让 handler 不依赖具体的数据库实现。
package service

import (
	"goproduct/domain/model"
	"goproduct/domain/repository"
	"goproduct/proto"
)

// IProductDataService 是一份方法清单。谁实现了这些方法，谁就可以被当成这个接口用。
// 名字前面的 I 是课程习惯。Go 社区更常见的是不加 I，直接叫 ProductDataService。
// 参数只写了类型、没写名字，是因为接口只关心「传进来什么、返回什么」，不关心参数叫什么。
type IProductDataService interface {
	Page(int32, int32) (count int64, products *[]model.Product, err error)
	ShowProductDetail(int32) (obj *model.ProductDetail, err error)
	ShowProductSku(int32) (obj *[]model.ProductSku, err error)
	ShowDetailSku(int32) (obj *model.ProductSku, err error)
	UpdateSku(req *proto.UpdateSkuReq) (isSuccess bool, err error)
	CountNum() int64
}

// ProductDataService 是上面接口的一种实现。小写开头的字段别的包看不见，只能通过方法使用。
type ProductDataService struct {
	productRepository repository.IProductRepository
}

// NewProductDataService 是构造函数。Go 没有 new 关键字的特殊语法，约定用 New 开头的函数代替。
// 返回的是接口，调用方只能使用接口里列出的方法。
func NewProductDataService(productRepository repository.IProductRepository) IProductDataService {
	return &ProductDataService{productRepository: productRepository}
}

// Page 不写 SQL，只把页码交给 repository。这种薄薄的一层是为了让以后能在这里加校验、缓存，而不改 handler。
func (u *ProductDataService) Page(length int32, pageIndex int32) (count int64, products *[]model.Product, err error) {

	return u.productRepository.Page(length, pageIndex)
}

func (u *ProductDataService) CountNum() int64 {

	return u.productRepository.CountNum()
}

func (u *ProductDataService) ShowProductDetail(id int32) (product *model.ProductDetail, err error) {

	return u.productRepository.ShowProductDetail(id)
}
func (u *ProductDataService) ShowProductSku(id int32) (product *[]model.ProductSku, err error) {

	return u.productRepository.ShowProductSku(id)
}
func (u *ProductDataService) ShowDetailSku(id int32) (product *model.ProductSku, err error) {

	return u.productRepository.ShowDetailSku(id)
}

func (u *ProductDataService) UpdateSku(req *proto.UpdateSkuReq) (isSuccess bool, err error) {

	return u.productRepository.UpdateSku(req)
}
