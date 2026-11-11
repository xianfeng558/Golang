-- 老马代驾的表，给人读。各服务启动时 AutoMigrate，不用手工执行。
-- 五个进程各有自己的库，乘客不直接改司机表。金额除计价和行程外不要和「元」混用，计价与行程用「分」。

-- 库 laomadj_customer

CREATE TABLE IF NOT EXISTS `customers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '乘客主键，写进 JWT 的 jti',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '有值时普通查询看不到',
  `telephone` varchar(15) DEFAULT NULL COMMENT '登录手机号，唯一',
  `name` varchar(191) DEFAULT NULL COMMENT '姓名，可空。空串会撞唯一索引，所以用 NULL',
  `email` varchar(191) DEFAULT NULL COMMENT '邮箱，可空',
  `wechat` varchar(191) DEFAULT NULL COMMENT '微信，可空',
  `city_id` bigint unsigned DEFAULT NULL COMMENT '城市，询价目前固定城市 1',
  `token` varchar(512) DEFAULT NULL COMMENT '当前 JWT。重新登录覆盖，退出清空',
  `token_created_at` datetime(3) DEFAULT NULL COMMENT '最近签发时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_customers_telephone` (`telephone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='乘客。验证码在 Redis 1 号库，键 CVC:手机号，60 秒';

CREATE TABLE IF NOT EXISTS `trip_orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '行程主键',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `customer_id` bigint unsigned DEFAULT NULL COMMENT '乘客主键',
  `origin` longtext COMMENT '起点，经度,纬度',
  `destination` longtext COMMENT '终点，经度,纬度',
  `price` bigint DEFAULT NULL COMMENT '预估价格，单位分',
  `status` varchar(32) DEFAULT NULL COMMENT 'wait_driver 待接单，ongoing 行程中，finished 已完成，canceled 已取消',
  PRIMARY KEY (`id`),
  KEY `idx_trip_orders_customer` (`customer_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程。询价不写这张表，乘客确认叫车才插入。不能从已完成再取消';

-- 库 laomadj_driver

CREATE TABLE IF NOT EXISTS `drivers` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '司机主键',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `telephone` varchar(16) DEFAULT NULL COMMENT '登录手机号，唯一。JWT 的 jti 是手机号，不是主键',
  `token` varchar(2047) DEFAULT NULL COMMENT '当前 JWT，退出写成 NULL',
  `status` enum('out','in','listen','stop') DEFAULT NULL COMMENT 'out 出车，in 收车，listen 听单，stop 停用。新登记是 stop',
  `name` varchar(255) DEFAULT NULL COMMENT '姓名',
  `id_number` char(18) DEFAULT NULL COMMENT '身份证，唯一，当前没有录入接口',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_drivers_telephone` (`telephone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司机。验证码在 Redis 2 号库，键 DVC:手机号。停用不能直接出车';

-- 库 laomadj_valuation

CREATE TABLE IF NOT EXISTS `price_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '规则主键，示例数据固定 1、2、3',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `city_id` bigint unsigned DEFAULT NULL COMMENT '城市，示例只有 1',
  `start_fee` bigint DEFAULT NULL COMMENT '起步价，单位分。含 5 公里',
  `distance_fee` bigint DEFAULT NULL COMMENT '超出起步里程后每公里的费用，单位分',
  `duration_fee` bigint DEFAULT NULL COMMENT '每分钟费用，单位分',
  `start_at` int DEFAULT NULL COMMENT '开始小时，包含',
  `end_at` int DEFAULT NULL COMMENT '结束小时，不包含。7 和 23 表示 7 点到 23 点前',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='计价规则。不足 1 公里或 1 分钟的零头不计费';
