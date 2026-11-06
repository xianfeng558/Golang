// Package model 放商品相关的表结构。GORM 靠结构体字段和 tag 生成列。
package model

import "time"

// Product 对应表 product，一行是一件商品，不包含具体规格。
// 字段名大写，别的包（handler、repository）才能使用。
// json 标签决定接口返回时的键名；gorm 标签决定建表时的默认值。
// 名为 ID 的字段会被 GORM 当成主键，列名自动变成 id。
type Product struct {
	ID                int32     `json:"id"`                           // 商品编号，数据库自增
	Name              string    `json:"name"`                         // 商品名称
	ProductType       int32     `gorm:"default:1" json:"productType"` // 商品类型，课程里 1 表示普通商品
	CategoryId        int32     `json:"categoryId"`                   // 分类编号
	StartingPrice     float32   `json:"startingPrice"`                // 起步价，列表页展示用
	TotalStock        int32     `gorm:"default:1234" json:"totalStock"`
	MainPicture       string    `gorm:"default:1" json:"mainPicture"` // 封面图地址
	RemoteAreaPostage float32   `json:"remoteAreaPostage"`            // 偏远地区邮费
	SingleBuyLimit    int32     `json:"singleBuyLimit"`               // 单次购买上限，0 表示不限制
	IsEnable          bool      `json:"isEnable"`                     // 是否上架。bool 在 MySQL 里通常是 tinyint
	Remark            string    `gorm:"default:1" json:"remark"`
	CreateUser        int32     `gorm:"default:1" json:"createUser"`
	CreateTime        time.Time `json:"createTime"` // time.Time 存成 datetime
	UpdateUser        int32     `json:"updateUser"`
	UpdateTime        time.Time `json:"updateTime"`
	IsDeleted         bool      `json:"isDeleted"` // 软删除标记。true 表示业务上已删除，行还在表里
}

// ProductDetail 不是一张表，而是详情 SQL 查出来之后装数据的形状。
// 它把商品主表、图文详情、图片列表拼在一个结构体里，再交给 handler 转成 proto。
type ProductDetail struct {
	//gorm.Model
	ID                int32     `json:"id"`
	Name              string    `json:"name"`
	ProductType       int32     `gorm:"default:1" json:"productType"`
	CategoryId        int32     `json:"categoryId"`
	StartingPrice     float32   `json:"startingPrice"`
	TotalStock        int32     `gorm:"default:1234" json:"totalStock"`
	MainPicture       string    `gorm:"default:1" json:"mainPicture"`
	RemoteAreaPostage float32   `json:"remoteAreaPostage"`
	SingleBuyLimit    int32     `json:"singleBuyLimit"`
	IsEnable          bool      `json:"isEnable"`
	Remark            string    `gorm:"default:1" json:"remark"`
	CreateUser        int32     `gorm:"default:1" json:"createUser"`
	CreateTime        time.Time `json:"createTime"`
	UpdateUser        int32     `json:"updateUser"`
	UpdateTime        time.Time `json:"updateTime"`
	IsDeleted         bool      `json:"isDeleted"`
	Detail            string    `gorm:"detail" json:"detail"`            //商品详情页面
	PictureList       string    `gorm:"picture_list" json:"pictureList"` //商品详情需要的图片
}

// TableName 指定真实表名。不写的话 GORM 会把 Product 变成复数 products。
func (table *Product) TableName() string {
	return "product"
}

// ProductDetailRow 对应 product_detail 表，一行是一段图文详情。
// 详情接口用 SQL 把它和商品主表拼在一起，不直接把这个结构体返回给前端。
type ProductDetailRow struct {
	ID        int32  `gorm:"primaryKey" json:"id"`
	ProductId int32  `gorm:"column:product_id" json:"productId"`
	Detail    string `gorm:"column:detail" json:"detail"`
}

func (ProductDetailRow) TableName() string { return "product_detail" }

// ProductPicture 对应 product_picture 表，一个商品可以有多张图。
type ProductPicture struct {
	ID        int32  `gorm:"primaryKey" json:"id"`
	ProductId int32  `gorm:"column:product_id" json:"productId"`
	Picture   string `gorm:"column:picture" json:"picture"`
}

func (ProductPicture) TableName() string { return "product_picture" }
