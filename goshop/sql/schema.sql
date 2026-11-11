-- GoShop 表结构对照。进程启动时用 AutoMigrate 建表，这份文件用来读，也用来面试时讲表为什么这样拆。
-- 三个库：user_center（用户）、product_center（商品和购物车）、trade（订单）。
-- 购物车和商品规格必须在同一个库，算价 SQL 要 JOIN product_sku。

-- ---------- user_center ----------

CREATE TABLE IF NOT EXISTS `user` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '用户主键。登录后请求头 Uuid 可以填这个十进制数，或它的二进制形式',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像地址',
  `client_id` int DEFAULT 1 COMMENT '客户端编号。登录接口要求传 0',
  `nickname` varchar(255) DEFAULT '随机名称' COMMENT '昵称',
  `phone` varchar(32) DEFAULT NULL COMMENT '登录手机号。验证码固定 6666 时，没有就插入一行',
  `password` varchar(64) DEFAULT '1234' COMMENT '当前登录不校验密码',
  `system_id` varchar(32) DEFAULT '1' COMMENT '系统编号。登录接口要求传 0',
  `last_login_time` datetime DEFAULT NULL COMMENT '最近登录时间',
  `create_time` datetime DEFAULT NULL COMMENT '注册时间',
  `is_deleted` int DEFAULT 0 COMMENT '0 未删除',
  `union_id` varchar(64) DEFAULT '1' COMMENT '预留的联合编号',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户。令牌不在这张表，而在 Redis，键是用户 ID，过期 1 小时';

CREATE TABLE IF NOT EXISTS `user_address` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '地址主键',
  `user_id` int NOT NULL COMMENT '用户 ID',
  `receiver` varchar(64) NOT NULL COMMENT '收货人',
  `phone` varchar(16) NOT NULL COMMENT '收货手机，11 位',
  `province` varchar(32) DEFAULT NULL COMMENT '省',
  `city` varchar(32) DEFAULT NULL COMMENT '市',
  `detail` varchar(255) NOT NULL COMMENT '详细地址',
  `is_default` int DEFAULT 0 COMMENT '1 默认地址',
  `is_deleted` int DEFAULT 0 COMMENT '1 已删除',
  PRIMARY KEY (`id`),
  KEY `idx_user_address_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='收货地址。一个用户多条，校验规则在 user-service/domain/model/address.go';

-- ---------- product_center ----------

CREATE TABLE IF NOT EXISTS `product_category` (
  `id` int NOT NULL COMMENT '分类主键。示例商品使用 31',
  `parent_id` int DEFAULT 0 COMMENT '父分类，0 是顶级',
  `name` varchar(64) NOT NULL COMMENT '分类名',
  `sort` int DEFAULT 0 COMMENT '排序，小的靠前',
  `is_enable` tinyint(1) DEFAULT 1 COMMENT '是否展示',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品分类。商品启动时会写入数码、图书、学习示例';

CREATE TABLE IF NOT EXISTS `product` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '商品主键，不是规格',
  `name` varchar(255) DEFAULT NULL COMMENT '商品名',
  `product_type` int DEFAULT 1 COMMENT '1 普通商品',
  `category_id` int DEFAULT NULL COMMENT '分类 ID',
  `starting_price` float DEFAULT NULL COMMENT '列表展示价，单位元',
  `total_stock` int DEFAULT 1234 COMMENT '展示用总库存。真正扣减的是 product_sku.stock',
  `main_picture` varchar(512) DEFAULT NULL COMMENT '封面',
  `is_enable` tinyint(1) DEFAULT NULL COMMENT '是否上架',
  `is_deleted` tinyint(1) DEFAULT NULL COMMENT '业务删除标记，行仍在',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品主表。规格、详情、图片分表，避免一行里塞多张图';

CREATE TABLE IF NOT EXISTS `product_sku` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '规格主键，加购时传的 productSkuId',
  `product_id` int DEFAULT NULL COMMENT '所属商品',
  `name` varchar(255) DEFAULT NULL COMMENT '规格名',
  `attribute_symbol_list` varchar(255) DEFAULT NULL COMMENT '规格属性文字',
  `sell_price` float DEFAULT NULL COMMENT '售价，单位元',
  `stock` int DEFAULT 1 COMMENT '可售库存。加购用 stock = stock + delta，并且结果不能小于 0',
  PRIMARY KEY (`id`),
  KEY `idx_product_sku_product` (`product_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='可售规格。并发扣库存不在网关内存里做';

CREATE TABLE IF NOT EXISTS `product_detail` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '详情主键',
  `product_id` int DEFAULT NULL COMMENT '商品 ID',
  `detail` text COMMENT '图文详情 HTML',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品图文详情';

CREATE TABLE IF NOT EXISTS `product_picture` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '图片主键',
  `product_id` int DEFAULT NULL COMMENT '商品 ID',
  `picture` varchar(512) DEFAULT NULL COMMENT '图片地址',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='商品轮播图，一个商品多行';

CREATE TABLE IF NOT EXISTS `shopping_cart` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '购物车主键。走 Saga 时由网关事先生成，补偿请求才能带上它',
  `user_id` int DEFAULT 1 COMMENT '用户 ID',
  `product_id` int DEFAULT NULL COMMENT '商品 ID',
  `product_sku_id` int DEFAULT NULL COMMENT '规格 ID',
  `product_name` varchar(255) DEFAULT NULL COMMENT '加入时的商品名快照',
  `product_main_picture` varchar(512) DEFAULT NULL COMMENT '加入时的封面快照',
  `number` int DEFAULT 1 COMMENT '数量',
  `is_deleted` tinyint(1) DEFAULT 0 COMMENT '下单后标成已删除。false 也必须能写回，否则订单补偿无法恢复',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='购物车。和 product_sku 同库，才能按数量乘售价';

-- ---------- trade ----------

CREATE TABLE IF NOT EXISTS `trade_order` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '订单头主键',
  `order_no` varchar(64) DEFAULT NULL COMMENT '订单号，Y 加时间加用户加随机数，给支付回调对账',
  `user_id` int DEFAULT NULL COMMENT '下单用户',
  `total_amount` float DEFAULT NULL COMMENT '商品总额，单位元',
  `pay_amount` float DEFAULT NULL COMMENT '应付金额',
  `order_status` int DEFAULT 1 COMMENT '1待支付 2已关闭 3已支付 4已发货 5已收货 6已完成 7已追评',
  `is_deleted` tinyint(1) DEFAULT 0 COMMENT 'Saga 补偿时标删除',
  `submit_time` datetime DEFAULT NULL COMMENT '提交时间',
  `expire_time` datetime DEFAULT NULL COMMENT '支付截止，创建后 30 分钟',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单头。行项目在 trade_order_item，状态变化在 trade_order_log';

CREATE TABLE IF NOT EXISTS `trade_order_item` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '订单行',
  `order_id` int NOT NULL COMMENT '订单头',
  `cart_id` int DEFAULT 0 COMMENT '来源购物车。交易库不能 JOIN 商品库，所以只记 ID',
  `title` varchar(255) DEFAULT NULL COMMENT '下单时的名称快照',
  `quantity` int DEFAULT 1 COMMENT '数量',
  `pay_amount` float DEFAULT 0 COMMENT '这一行的金额。按购物车拆开时金额留在订单头',
  `create_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_trade_item_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单行。一张订单头对应多行';

CREATE TABLE IF NOT EXISTS `trade_order_log` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '状态流水',
  `order_id` int NOT NULL COMMENT '订单头',
  `from_status` int DEFAULT 0 COMMENT '变化前，新建时是 0',
  `to_status` int NOT NULL COMMENT '变化后',
  `remark` varchar(255) DEFAULT NULL COMMENT '创建订单或更新订单',
  `create_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_trade_log_order` (`order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单状态流水。不允许从已发货跳回待支付';
