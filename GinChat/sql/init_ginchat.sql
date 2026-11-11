-- 这份 SQL 用来对照表结构，给人读。进程启动时会 AutoMigrate，不用手工执行。
-- 库名是 ginchat，字符集 utf8mb4。
-- user_basic 是用户。contact 是已经成立的好友和群关系，type=1 好友，type=2 群成员。
-- friend_apply 是还没成立的好友申请。communities 是当前页面使用的群。
-- message 的列和模型一致，但私聊记录实际写在 Redis。group_basic 是旧表，当前建群不写它。

CREATE TABLE IF NOT EXISTS `user_basic` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `pass_word` longtext,
  `phone` longtext,
  `email` longtext,
  `avatar` longtext,
  `identity` longtext,
  `client_ip` longtext,
  `client_port` longtext,
  `salt` longtext,
  `login_time` datetime(3) DEFAULT NULL,
  `heartbeat_time` datetime(3) DEFAULT NULL,
  `login_out_time` datetime(3) DEFAULT NULL,
  `is_logout` tinyint(1) DEFAULT NULL,
  `device_info` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_user_basic_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户。密码是 MD5(明文+盐)，令牌在 identity，WebSocket 的 token 必须等于它';

CREATE TABLE IF NOT EXISTS `contact` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `owner_id` bigint unsigned DEFAULT NULL,
  `target_id` bigint unsigned DEFAULT NULL,
  `type` bigint DEFAULT NULL,
  `desc` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_contact_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='已成立的关系。好友两行，群成员一行。type=1 好友，type=2 群';

CREATE TABLE IF NOT EXISTS `communities` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `owner_id` bigint unsigned DEFAULT NULL,
  `img` longtext,
  `desc` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_communities_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群。建群时群主会同时得到一条 type=2 的 contact';

CREATE TABLE IF NOT EXISTS `friend_apply` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '申请主键',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `from_id` bigint unsigned DEFAULT NULL COMMENT '申请人',
  `to_id` bigint unsigned DEFAULT NULL COMMENT '被申请人，只有这个人能同意或拒绝',
  `status` bigint DEFAULT 0 COMMENT '0 待处理，1 已同意，2 已拒绝',
  `remark` longtext COMMENT '申请附言',
  PRIMARY KEY (`id`),
  KEY `idx_friend_apply_deleted_at` (`deleted_at`),
  KEY `idx_friend_apply_to` (`to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友申请。同意后才写 contact。/contact/addfriend 仍然直接成为好友';

CREATE TABLE IF NOT EXISTS `message` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `user_id` bigint DEFAULT NULL,
  `target_id` bigint DEFAULT NULL,
  `type` bigint DEFAULT NULL,
  `media` bigint DEFAULT NULL,
  `content` longtext,
  `create_time` bigint unsigned DEFAULT NULL,
  `read_time` bigint unsigned DEFAULT NULL,
  `pic` longtext,
  `url` longtext,
  `desc` longtext,
  `amount` bigint DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_message_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表结构。当前私聊正文在 Redis 的 msg_较小ID_较大ID，不插这张表';

CREATE TABLE IF NOT EXISTS `group_basic` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  `deleted_at` datetime(3) DEFAULT NULL,
  `name` longtext,
  `owner_id` bigint unsigned DEFAULT NULL,
  `icon` longtext,
  `type` bigint DEFAULT NULL,
  `desc` longtext,
  PRIMARY KEY (`id`),
  KEY `idx_group_basic_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
