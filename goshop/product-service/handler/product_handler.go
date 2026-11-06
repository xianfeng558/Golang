// Package handler 实现商品的五个 RPC：分页、详情、规格列表、单个规格、改库存。
// req 是调用方传来的参数，resp 由本方法填好后由框架送回。ctx 这里没用到，但签名必须和生成代码一致。
package handler

import (
	"context"
	"errors"

	common "git.mashibing.com/msb_47094/shopping-comm"
	"goproduct/domain/model"
	"goproduct/domain/service"
	"goproduct/proto"
)

// ProductHandler 挂在 RPC 服务器上。字段用接口，所以换数据库实现时不用改这里。
type ProductHandler struct {
	ProductDataService service.IProductDataService
}

// Page 按页返回商品。查询失败直接返回错误，避免后面拿到空指针。
func (u *ProductHandler) Page(ctx context.Context, req *proto.PageReq, resp *proto.PageResp) error {
	count, obj, err := u.ProductDataService.Page(req.GetLength(), req.GetPageIndex())
	if err != nil {
		return err
	}
	resp.Rows = int64(req.GetLength())
	resp.Total = count
	return copyProducts(obj, resp)
}

// copyProducts 把数据库里的商品逐条抄进 proto。*[]Product 是切片的指针，遍历前要先 *obj。
func copyProducts(obj *[]model.Product, resp *proto.PageResp) error {
	if obj == nil {
		return nil
	}
	for _, v := range *obj {
		product := &proto.Product{}
		if err := common.SwapToStruct(v, product); err != nil {
			return err
		}
		resp.Product = append(resp.Product, product)
	}
	return nil
}

// ShowProductDetail 按商品 ID 查详情。查不到就返回错误，调用方会看到「请求失败」。
func (u *ProductHandler) ShowProductDetail(ctx context.Context, req *proto.ProductDetailReq, resp *proto.ProductDetailResp) error {
	obj, err := u.ProductDataService.ShowProductDetail(req.GetId())
	if err != nil || obj == nil {
		if err == nil {
			err = errors.New("商品不存在")
		}
		return err
	}
	productDetail := &proto.ProductDetail{}
	if err = common.SwapToStruct(obj, productDetail); err != nil {
		return err
	}
	resp.ProductDetail = append(resp.ProductDetail, productDetail)
	return nil
}

// ShowProductSku 返回某个商品下的全部规格。一件商品可以有多条 SKU。
func (u *ProductHandler) ShowProductSku(ctx context.Context, req *proto.ProductSkuReq, resp *proto.ProductSkuResp) error {
	obj, err := u.ProductDataService.ShowProductSku(req.GetProductId())
	if err != nil {
		return err
	}
	return copySkus(obj, resp)
}

func copySkus(obj *[]model.ProductSku, resp *proto.ProductSkuResp) error {
	if obj == nil {
		return nil
	}
	for _, v := range *obj {
		productSku := &proto.ProductSku{}
		if err := common.SwapToStruct(v, productSku); err != nil {
			return err
		}
		resp.ProductSku = append(resp.ProductSku, productSku)
	}
	return nil
}

// ShowDetailSku 按规格 ID 查一条库存。加购前用 Stock 判断够不够。
func (u *ProductHandler) ShowDetailSku(ctx context.Context, req *proto.ProductDetailReq, resp *proto.ProductSkuResp) error {
	obj, err := u.ProductDataService.ShowDetailSku(req.GetId())
	if err != nil {
		return err
	}
	productSku := &proto.ProductSku{}
	if err = common.SwapToStruct(obj, productSku); err != nil {
		return err
	}
	resp.ProductSku = append(resp.ProductSku, productSku)
	return nil
}

// UpdateSku 修改规格库存。请求里没有规格数据时直接失败。
func (u *ProductHandler) UpdateSku(ctx context.Context, req *proto.UpdateSkuReq, resp *proto.UpdateSkuResp) error {
	if req.GetProductSku() == nil {
		resp.IsSuccess = false
		return errors.New("productSku 为空")
	}
	isSuccess, err := u.ProductDataService.UpdateSku(req)
	resp.IsSuccess = isSuccess
	return err
}
