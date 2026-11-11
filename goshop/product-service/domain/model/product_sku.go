package model

// ProductSku 是一件商品下「可以卖的具体规格」，对应表 product_sku。
// 例如同一本书可以有「平装」「精装」两个 SKU，库存和价格各记各的。
//
// 反引号里的内容叫结构体标签，Go 自己不解释，是给别的库看的：
//   - gorm:"..." 告诉 GORM 这一列在 MySQL 里叫什么、是不是主键
//   - json:"..." 告诉 encoding/json 编成 JSON 时用哪个名字
//
// 字段名要大写才能被别的包访问。小写字段只在本包内可见。
type ProductSku struct {
	// primaryKey 主键；column:id 列名是 id，不是 sku_id；autoIncrement 由数据库生成编号。
	SkuId int32 `gorm:"primaryKey;column:id;autoIncrement;comment:规格主键" json:"skuId"`
	// 这条规格属于哪件商品。详情和加购都靠它关联 product 表。
	ProductId int32 `gorm:"column:product_id;comment:所属商品 ID" json:"productId"`
	// 规格名称，例如「默认规格」。没有 gorm 标签时，列名按字段名转成下划线：name。
	Name string `gorm:"comment:规格名称" json:"name"`
	// 规格属性的文字描述。column 把 Go 的大驼峰映射成表里的下划线列名。
	AttributeSymbolList string `gorm:"column:attribute_symbol_list;comment:规格属性文字" json:"attributeSymbolList"`
	// float32 是 32 位小数。金额用浮点数会有精度问题，这里沿用课程字段，正式项目常用整数分。
	SellPrice float32 `gorm:"column:sell_price;comment:售价，单位元" json:"sellPrice"`
	// default:1 只在建表时给列一个默认值。插入时如果代码写了 Stock，以代码为准。
	Stock int32 `gorm:"default:1;comment:可售库存，加减库存改这一列" json:"stock"`
}

func (ProductSku) TableComment() string { return "可售规格，库存按这一行加减" }

// TableName 是 GORM 约定的方法名。返回值就是真实表名。
// 不写的话，GORM 会把 ProductSku 变成复数 product_skus，和课程里的表对不上。
// (table *ProductSku) 是接收者：这个函数属于 ProductSku，table 相当于别的语言里的 this。
// 指针接收者表示方法里可以修改原对象；这里只读表名，用不用指针效果一样，和同包其他模型保持一致。
func (table *ProductSku) TableName() string {
	return "product_sku"
}
