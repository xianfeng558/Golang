// Package service 把商品查询和改库存转交给 repository。
// 名字前面的 I 是课程习惯。接口只写类型、不写参数名，因为这里只关心传入和返回。
package service

import (
	"goproduct/domain/model"
	"goproduct/domain/repository"
	"goproduct/proto"
)

// IProductDataService 是 handler 能调用的方法清单。
type IProductDataService interface {
	Page(int32, int32) (count int64, products *[]model.Product, err error)
	ShowProductDetail(int32) (obj *model.ProductDetail, err error)
	ShowProductSku(int32) (obj *[]model.ProductSku, err error)
	ShowDetailSku(int32) (obj *model.ProductSku, err error)
	UpdateSku(req *proto.UpdateSkuReq) (isSuccess bool, err error)
}

// ProductDataService 是上面接口的实现。小写字段只能在本包里使用。
type ProductDataService struct {
	productRepository repository.IProductRepository
}

// NewProductDataService 是构造函数。Go 没有专门的构造语法，约定用 New 开头，并返回接口。
func NewProductDataService(productRepository repository.IProductRepository) IProductDataService {
	return &ProductDataService{productRepository: productRepository}
}

func (u *ProductDataService) Page(length int32, pageIndex int32) (int64, *[]model.Product, error) {
	return u.productRepository.Page(length, pageIndex)
}

func (u *ProductDataService) ShowProductDetail(id int32) (*model.ProductDetail, error) {
	return u.productRepository.ShowProductDetail(id)
}

func (u *ProductDataService) ShowProductSku(id int32) (*[]model.ProductSku, error) {
	return u.productRepository.ShowProductSku(id)
}

func (u *ProductDataService) ShowDetailSku(id int32) (*model.ProductSku, error) {
	return u.productRepository.ShowDetailSku(id)
}

func (u *ProductDataService) UpdateSku(req *proto.UpdateSkuReq) (bool, error) {
	return u.productRepository.UpdateSku(req)
}
