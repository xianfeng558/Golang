// Package handler 实现商品 proto 里的五个 RPC：分页、详情、规格列表、单个规格、改库存。
package handler

import (
	"context"
	"errors"
	"fmt"
	comm "goproduct/common"
	"goproduct/domain/model"
	"goproduct/domain/service"
	"goproduct/proto"
)

// ProductHandler 是挂到 RPC 服务器上的对象。
// 字段是接口而不是具体结构体：只要有 Page、ShowProductDetail 这些方法，就可以塞进来。
// 这样 handler 不关心数据存在 MySQL 还是别的地方。
type ProductHandler struct {
	ProductDataService service.IProductDataService
}

// Page 对应 product.proto 里的 rpc Page。
// (u *ProductHandler) 表示这是 ProductHandler 的方法，u 就是当前这个对象。
// req 是对方传来的参数，resp 是我们填好后由框架送回去的结果。返回 error 表示这次调用失败。
// ctx 本方法没有使用，但签名必须和生成代码一致，否则注册不上去。
func (u *ProductHandler) Page(ctx context.Context, req *proto.PageReq, resp *proto.PageResp) error {
	//count := u.ProductDataService.CountNum()
	count, obj, err := u.ProductDataService.Page(req.GetLength(), req.GetPageIndex())
	if err != nil {
		println("page product err :", err)
	}
	//count = u.ProductDataService.CountNum()
	resp.Rows = int64(req.GetLength())
	resp.Total = count
	//fmt.Println(">>>>>>>>>>>>> page product success :", obj)
	ObjForResp(obj, resp)
	return nil
}

// ObjForResp 把数据库里的 []Product 逐条抄进 proto 的列表。
// *[]model.Product 是「切片的指针」。调用时要先 *obj 才能遍历里面的元素。
// append 向切片尾部追加元素。切片容量不够时，Go 会自动换一块更大的内存。
func ObjForResp(obj *[]model.Product, resp *proto.PageResp) (err error) {
	for _, v := range *obj {
		product := &proto.Product{}
		err := comm.SwapToStruct(v, product)
		if err != nil {
			return err
		}
		fmt.Println(">>>>>>>>>>>>> ", product)
		resp.Product = append(resp.Product, product)
	}
	return nil
}

// ShowProductDetail 按商品 ID 查详情。查不到就返回 error，调用方会看到「请求失败」，不会拿到空商品。
func (u *ProductHandler) ShowProductDetail(ctx context.Context, req *proto.ProductDetailReq, resp *proto.ProductDetailResp) error {
	//count := u.ProductDataService.CountNum()
	obj, err := u.ProductDataService.ShowProductDetail(req.GetId())
	if err != nil || obj == nil {
		println("ShowProductDetail   err :", err)
		return err
	}
	productDetail := &proto.ProductDetail{}
	err1 := comm.SwapToStruct(obj, productDetail)
	if err1 != nil {
		println("ShowProductDetail SwapToStruct  err :", err1)
	}
	resp.ProductDetail = append(resp.ProductDetail, productDetail)
	return nil
}

// ShowProductSku 返回某个商品下的全部规格。一件商品可以有多条 SKU，所以结果是列表。
func (u *ProductHandler) ShowProductSku(ctx context.Context, req *proto.ProductSkuReq, resp *proto.ProductSkuResp) error {
	//count := u.ProductDataService.CountNum()
	obj, err := u.ProductDataService.ShowProductSku(req.GetProductId())
	if err != nil {
		println("ShowProductSku   err :", err)
	}
	err1 := ObjSkuForResp(obj, resp)
	if err1 != nil {
		println("ShowProductSku SwapToStruct  err :", err1)
	}
	return nil
}

// ObjSkuForResp 和 ObjForResp 一样，只是元素从商品换成了规格。
func ObjSkuForResp(obj *[]model.ProductSku, resp *proto.ProductSkuResp) (err error) {
	for _, v := range *obj {
		productSku := &proto.ProductSku{}
		err := comm.SwapToStruct(v, productSku)
		if err != nil {
			return err
		}
		resp.ProductSku = append(resp.ProductSku, productSku)
	}
	return nil
}

// ShowDetailSku 按规格 ID 查一条库存。加购前用它判断 Stock 够不够。
func (u *ProductHandler) ShowDetailSku(ctx context.Context, req *proto.ProductDetailReq, resp *proto.ProductSkuResp) error {
	//count := u.ProductDataService.CountNum()
	obj, err := u.ProductDataService.ShowDetailSku(req.Id)
	if err != nil {
		println("ShowDetailSku   err :", err)
	}
	productSku := &proto.ProductSku{}
	err = comm.SwapToStruct(obj, productSku)
	if err != nil {
		return err
	}
	resp.ProductSku = append(resp.ProductSku, productSku)
	return nil
}

// UpdateSku 修改规格库存。请求里没有规格数据时直接失败，避免后面空指针。
func (u *ProductHandler) UpdateSku(ctx context.Context, req *proto.UpdateSkuReq, resp *proto.UpdateSkuResp) error {
	if req.GetProductSku() == nil {
		resp.IsSuccess = false
		return errors.New("productSku 为空")
	}
	isSuccess, err := u.ProductDataService.UpdateSku(req)
	if err != nil {
		resp.IsSuccess = isSuccess
		println("UpdateSku   err :", err)
	}
	resp.IsSuccess = isSuccess
	return err
}
