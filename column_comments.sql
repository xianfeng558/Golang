-- 给现有业务表补表注释和列注释。只改 COMMENT，类型、空值、默认值和自增保持原样。
-- 不碰 activelife_godohosting_com 和 shaoyin_shopping_mall。

ALTER TABLE ginchat.communities COMMENT = '群';
ALTER TABLE ginchat.communities
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN name longtext NULL COMMENT '群名称',
  MODIFY COLUMN owner_id bigint(20) unsigned NULL COMMENT '群主用户 ID',
  MODIFY COLUMN img longtext NULL COMMENT '群头像地址',
  MODIFY COLUMN `desc` longtext NULL COMMENT '群简介';

ALTER TABLE ginchat.contact COMMENT = '已成立的好友或群关系';
ALTER TABLE ginchat.contact
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN owner_id bigint(20) unsigned NULL COMMENT '这条关系属于谁',
  MODIFY COLUMN target_id bigint(20) unsigned NULL COMMENT '好友时是对方用户 ID，加群时是群 ID',
  MODIFY COLUMN type bigint(20) NULL COMMENT '1 好友，2 群成员',
  MODIFY COLUMN `desc` longtext NULL COMMENT '备注';

ALTER TABLE ginchat.friend_apply COMMENT = '好友申请，同意后才写入 contact';
ALTER TABLE ginchat.friend_apply
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN from_id bigint(20) unsigned NULL COMMENT '申请人用户 ID',
  MODIFY COLUMN to_id bigint(20) unsigned NULL COMMENT '被申请人用户 ID',
  MODIFY COLUMN status bigint(20) NULL COMMENT '0 待处理，1 已同意，2 已拒绝',
  MODIFY COLUMN remark longtext NULL COMMENT '申请附言';

ALTER TABLE ginchat.group_basic COMMENT = '旧群表，当前建群不写它';
ALTER TABLE ginchat.group_basic
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN name longtext NULL COMMENT '群名称',
  MODIFY COLUMN owner_id bigint(20) unsigned NULL COMMENT '群主用户 ID',
  MODIFY COLUMN icon longtext NULL COMMENT '群图标地址',
  MODIFY COLUMN type bigint(20) NULL COMMENT '预留的群分类',
  MODIFY COLUMN `desc` longtext NULL COMMENT '群说明';

ALTER TABLE ginchat.message COMMENT = '消息表结构，私聊正文实际在 Redis';
ALTER TABLE ginchat.message
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN user_id bigint(20) NULL COMMENT '发送者用户 ID',
  MODIFY COLUMN target_id bigint(20) NULL COMMENT '私聊时是对方用户 ID，群聊时是群 ID',
  MODIFY COLUMN type bigint(20) NULL COMMENT '1 私聊，2 群聊，3 心跳',
  MODIFY COLUMN media bigint(20) NULL COMMENT '1 文字，2 表情，3 语音，4 图片',
  MODIFY COLUMN content longtext NULL COMMENT '文字内容',
  MODIFY COLUMN create_time bigint(20) unsigned NULL COMMENT '页面带来的时间',
  MODIFY COLUMN read_time bigint(20) unsigned NULL COMMENT '预留的已读时间',
  MODIFY COLUMN pic longtext NULL COMMENT '预留图片字段',
  MODIFY COLUMN url longtext NULL COMMENT '语音或图片地址',
  MODIFY COLUMN `desc` longtext NULL COMMENT '预留说明',
  MODIFY COLUMN amount bigint(20) NULL COMMENT '语音时长，单位秒';

ALTER TABLE ginchat.user_basic COMMENT = '注册用户';
ALTER TABLE ginchat.user_basic
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN name longtext NULL COMMENT '登录名',
  MODIFY COLUMN pass_word longtext NULL COMMENT '密码摘要，MD5(明文加盐)',
  MODIFY COLUMN phone longtext NULL COMMENT '手机号',
  MODIFY COLUMN email longtext NULL COMMENT '邮箱',
  MODIFY COLUMN avatar longtext NULL COMMENT '头像地址',
  MODIFY COLUMN identity longtext NULL COMMENT '登录令牌，WebSocket 的 token 必须等于它',
  MODIFY COLUMN client_ip longtext NULL COMMENT '预留的客户端 IP',
  MODIFY COLUMN client_port longtext NULL COMMENT '预留的客户端端口',
  MODIFY COLUMN salt longtext NULL COMMENT '6 位数字盐',
  MODIFY COLUMN login_time datetime(3) NULL COMMENT '最近一次签发令牌的时间',
  MODIFY COLUMN heartbeat_time datetime(3) NULL COMMENT '注册时记下的时间，连接心跳不回写这一列',
  MODIFY COLUMN login_out_time datetime(3) NULL COMMENT '预留的退出时间',
  MODIFY COLUMN is_logout tinyint(1) NULL COMMENT '预留的登出标记',
  MODIFY COLUMN device_info longtext NULL COMMENT '预留的设备信息';

ALTER TABLE laomadj_customer.customers COMMENT = '乘客。验证码在 Redis 1 号库，键是 CVC 加手机号';
ALTER TABLE laomadj_customer.customers
  MODIFY COLUMN telephone varchar(15) NULL COMMENT '登录手机号，唯一',
  MODIFY COLUMN name varchar(191) NULL COMMENT '姓名，可空',
  MODIFY COLUMN email varchar(191) NULL COMMENT '邮箱，可空',
  MODIFY COLUMN wechat varchar(191) NULL COMMENT '微信号，可空',
  MODIFY COLUMN city_id bigint(20) unsigned NULL COMMENT '所在城市',
  MODIFY COLUMN token varchar(512) NULL COMMENT '当前 JWT，重新登录会覆盖',
  MODIFY COLUMN token_created_at datetime(3) NULL COMMENT '最近一次签发时间',
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '乘客主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到';

ALTER TABLE laomadj_driver.drivers COMMENT = '司机。验证码在 Redis 2 号库，键是 DVC 加手机号';
ALTER TABLE laomadj_driver.drivers
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '司机主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN telephone varchar(16) NULL COMMENT '登录手机号，唯一',
  MODIFY COLUMN token varchar(2047) NULL COMMENT '当前 JWT，退出后为 NULL',
  MODIFY COLUMN status enum('out','in','listen','stop') NULL COMMENT 'out 出车，in 收车，listen 听单，stop 停用',
  MODIFY COLUMN name varchar(255) NULL COMMENT '姓名',
  MODIFY COLUMN id_number char(18) NULL COMMENT '身份证号，唯一',
  MODIFY COLUMN id_image_a varchar(255) NULL COMMENT '身份证人像面地址',
  MODIFY COLUMN license_image_a varchar(255) NULL COMMENT '驾驶证正本地址',
  MODIFY COLUMN license_image_b varchar(255) NULL COMMENT '驾驶证副本地址',
  MODIFY COLUMN distinct_code varchar(16) NULL COMMENT '行政区划代码',
  MODIFY COLUMN telephone_bak varchar(16) NULL COMMENT '备用手机号',
  MODIFY COLUMN audit_at datetime(3) NULL COMMENT '审核通过时间';

ALTER TABLE laomadj_valuation.price_rules COMMENT = '计价规则，金额单位是分';
ALTER TABLE laomadj_valuation.price_rules
  MODIFY COLUMN id bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '计价规则主键',
  MODIFY COLUMN created_at datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN updated_at datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN deleted_at datetime(3) NULL COMMENT '软删除时间，有值时普通查询看不到',
  MODIFY COLUMN city_id bigint(20) unsigned NULL COMMENT '城市编号',
  MODIFY COLUMN start_fee bigint(20) NULL COMMENT '起步价，单位分，含 5 公里',
  MODIFY COLUMN distance_fee bigint(20) NULL COMMENT '超出起步里程后每公里费用，单位分',
  MODIFY COLUMN duration_fee bigint(20) NULL COMMENT '每分钟费用，单位分',
  MODIFY COLUMN start_at bigint(20) NULL COMMENT '开始小时，包含',
  MODIFY COLUMN end_at bigint(20) NULL COMMENT '结束小时，不包含';

ALTER TABLE product_center.product COMMENT = '商品主表，不含具体规格';
ALTER TABLE product_center.product
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '商品主键',
  MODIFY COLUMN name longtext NULL COMMENT '商品名称',
  MODIFY COLUMN product_type int(11) NULL DEFAULT 1 COMMENT '1 普通商品',
  MODIFY COLUMN category_id int(11) NULL COMMENT '分类编号，对应 product_category',
  MODIFY COLUMN starting_price float NULL COMMENT '起步价，列表页展示',
  MODIFY COLUMN total_stock int(11) NULL DEFAULT 1234 COMMENT '展示用总库存，扣库存改的是 SKU',
  MODIFY COLUMN main_picture varchar(191) NULL DEFAULT '1' COMMENT '封面图地址',
  MODIFY COLUMN remote_area_postage float NULL COMMENT '偏远地区邮费',
  MODIFY COLUMN single_buy_limit int(11) NULL COMMENT '单次购买上限，0 表示不限制',
  MODIFY COLUMN is_enable tinyint(1) NULL COMMENT '是否上架',
  MODIFY COLUMN remark varchar(191) NULL DEFAULT '1' COMMENT '备注',
  MODIFY COLUMN create_user int(11) NULL DEFAULT 1 COMMENT '创建人',
  MODIFY COLUMN create_time datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN update_user int(11) NULL COMMENT '更新人',
  MODIFY COLUMN update_time datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN is_deleted tinyint(1) NULL COMMENT '软删除标记，true 表示业务上已删除';

ALTER TABLE product_center.product_detail COMMENT = '商品图文详情';
ALTER TABLE product_center.product_detail
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '详情主键',
  MODIFY COLUMN product_id int(11) NULL COMMENT '商品 ID',
  MODIFY COLUMN detail longtext NULL COMMENT '图文详情';

ALTER TABLE product_center.product_picture COMMENT = '商品图片，一个商品可以有多张';
ALTER TABLE product_center.product_picture
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '图片主键',
  MODIFY COLUMN product_id int(11) NULL COMMENT '商品 ID',
  MODIFY COLUMN picture longtext NULL COMMENT '图片地址';

ALTER TABLE product_center.product_sku COMMENT = '可售规格，库存按这一行加减';
ALTER TABLE product_center.product_sku
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '规格主键',
  MODIFY COLUMN name longtext NULL COMMENT '规格名称',
  MODIFY COLUMN attribute_symbol_list longtext NULL COMMENT '规格属性文字',
  MODIFY COLUMN sell_price float NULL COMMENT '售价，单位元',
  MODIFY COLUMN stock int(11) NULL DEFAULT 1 COMMENT '可售库存，加减库存改这一列',
  MODIFY COLUMN product_id int(11) NULL COMMENT '所属商品 ID';

ALTER TABLE product_center.shopping_cart COMMENT = '购物车';
ALTER TABLE product_center.shopping_cart
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '购物车主键',
  MODIFY COLUMN user_id int(11) NULL DEFAULT 1 COMMENT '用户 ID',
  MODIFY COLUMN product_id int(11) NULL COMMENT '商品 ID',
  MODIFY COLUMN product_sku_id int(11) NULL COMMENT '规格 ID',
  MODIFY COLUMN product_name longtext NULL COMMENT '加购时的商品名',
  MODIFY COLUMN product_main_picture longtext NULL COMMENT '加购时的封面',
  MODIFY COLUMN number int(11) NULL DEFAULT 1 COMMENT '数量',
  MODIFY COLUMN create_user int(11) NULL DEFAULT 1 COMMENT '创建人',
  MODIFY COLUMN create_time datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN update_user int(11) NULL COMMENT '更新人',
  MODIFY COLUMN update_time datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN is_deleted tinyint(1) NULL COMMENT '软删除标记';

ALTER TABLE trade_center.trade_order COMMENT = '订单头';
ALTER TABLE trade_center.trade_order
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '订单主键',
  MODIFY COLUMN order_no longtext NULL COMMENT '订单号',
  MODIFY COLUMN user_id int(11) NULL DEFAULT 1 COMMENT '下单用户 ID',
  MODIFY COLUMN total_amount float NULL COMMENT '商品总额，单位元',
  MODIFY COLUMN shipping_amount float NULL COMMENT '运费，单位元',
  MODIFY COLUMN discount_amount float NULL COMMENT '优惠金额，单位元',
  MODIFY COLUMN pay_amount float NULL COMMENT '应付金额，单位元',
  MODIFY COLUMN refund_amount float NULL COMMENT '退款金额，单位元',
  MODIFY COLUMN submit_time datetime(3) NULL COMMENT '提交时间',
  MODIFY COLUMN expire_time datetime(3) NULL COMMENT '支付截止时间',
  MODIFY COLUMN auto_receive_time datetime(3) NULL COMMENT '自动收货时间',
  MODIFY COLUMN receive_time datetime(3) NULL COMMENT '实际收货时间',
  MODIFY COLUMN auto_praise datetime(3) NULL COMMENT '自动好评时间',
  MODIFY COLUMN after_sale_deadline_time datetime(3) NULL COMMENT '售后截止时间',
  MODIFY COLUMN order_status int(11) NULL DEFAULT 1 COMMENT '1 待支付，2 已关闭，3 已支付，4 已发货，5 已收货，6 已完成，7 已追评',
  MODIFY COLUMN order_source int(11) NULL DEFAULT 6 COMMENT '订单来源',
  MODIFY COLUMN cancel_reason longtext NULL COMMENT '取消原因',
  MODIFY COLUMN order_type int(11) NULL DEFAULT 1 COMMENT '订单类型',
  MODIFY COLUMN create_user int(11) NULL DEFAULT 1 COMMENT '创建人',
  MODIFY COLUMN create_time datetime(3) NULL COMMENT '创建时间',
  MODIFY COLUMN update_user int(11) NULL COMMENT '更新人',
  MODIFY COLUMN update_time datetime(3) NULL COMMENT '更新时间',
  MODIFY COLUMN is_deleted tinyint(1) NULL COMMENT '软删除标记，回滚时要显式更新 false',
  MODIFY COLUMN pay_type int(11) NULL DEFAULT 1 COMMENT '支付方式',
  MODIFY COLUMN is_package_free int(11) NULL DEFAULT 1 COMMENT '是否包邮';

ALTER TABLE user_center.user COMMENT = '用户。令牌在 Redis，键是用户 ID';
ALTER TABLE user_center.user
  MODIFY COLUMN id int(11) NOT NULL AUTO_INCREMENT COMMENT '用户主键',
  MODIFY COLUMN avatar varchar(191) NULL DEFAULT 'https://msb-edu-dev.oss-cn-beijing.aliyuncs.com/default-headimg.png' COMMENT '头像地址',
  MODIFY COLUMN client_id int(11) NULL DEFAULT 1 COMMENT '客户端编号，登录接口要求传 0',
  MODIFY COLUMN nickname varchar(191) NULL DEFAULT '随机名称' COMMENT '昵称',
  MODIFY COLUMN phone longtext NULL COMMENT '登录手机号',
  MODIFY COLUMN password varchar(191) NULL DEFAULT '1234' COMMENT '当前登录不校验密码',
  MODIFY COLUMN system_id varchar(191) NULL DEFAULT '1' COMMENT '系统编号，登录接口要求传 0',
  MODIFY COLUMN last_login_time datetime(3) NULL COMMENT '最近登录时间',
  MODIFY COLUMN create_time datetime(3) NULL COMMENT '注册时间',
  MODIFY COLUMN is_deleted int(11) NULL DEFAULT 0 COMMENT '0 未删除',
  MODIFY COLUMN union_id varchar(191) NULL DEFAULT '1' COMMENT '预留的联合编号';
