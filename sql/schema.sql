-- =============================================================
-- go_backend_frame 数据库初始化脚本
-- 使用方式：手动在 MySQL 中执行本文件即可（无需迁移命令）
--   mysql -uroot -p < sql/schema.sql
-- 初始账号：
--   admin  / 123456  （超级管理员，拥有全部权限）
--   zhangsan / 123456 （运营专员，仅系统总览）
-- =============================================================

CREATE DATABASE IF NOT EXISTS `cf_backend_frame` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `cf_backend_frame`;
SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

-- ------------------------------------------------------------
-- 部门表（公司组织架构，多级）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_dept`;
CREATE TABLE `sys_dept` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `deleted_at` DATETIME(3) NULL DEFAULT NULL,
  `name`       VARCHAR(64)  NOT NULL COMMENT '部门名称',
  `parent_id`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上级部门ID，0=顶级',
  `sort`       INT NOT NULL DEFAULT 0 COMMENT '排序',
  `leader`     VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '负责人',
  `remark`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_sys_dept_deleted_at` (`deleted_at`),
  KEY `idx_sys_dept_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='部门表';

-- ------------------------------------------------------------
-- 用户表
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_user`;
CREATE TABLE `sys_user` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `deleted_at` DATETIME(3) NULL DEFAULT NULL,
  `username`   VARCHAR(32)  NOT NULL COMMENT '登录名',
  `password`   VARCHAR(128) NOT NULL COMMENT 'bcrypt密文',
  `nickname`   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '姓名',
	`avatar`     VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像相对路径',
  `mobile`     VARCHAR(16)  NOT NULL DEFAULT '' COMMENT '手机号',
  `email`      VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '邮箱',
  `dept_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属部门',
  `status`     TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 2禁用',
  `is_super`   TINYINT NOT NULL DEFAULT 0 COMMENT '1超级管理员',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  KEY `idx_sys_user_deleted_at` (`deleted_at`),
  KEY `idx_sys_user_dept_id` (`dept_id`),
  KEY `idx_sys_user_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户表';

-- ------------------------------------------------------------
-- 角色表（多级：上级角色自动拥有子级及后代角色的权限）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_role`;
CREATE TABLE `sys_role` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `deleted_at` DATETIME(3) NULL DEFAULT NULL,
  `name`       VARCHAR(32)  NOT NULL COMMENT '角色名称',
  `code`       VARCHAR(32)  NOT NULL COMMENT '角色编码',
  `parent_id`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上级角色ID，0=顶级',
  `sort`       INT NOT NULL DEFAULT 0 COMMENT '排序',
  `status`     TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 2禁用',
  `remark`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_sys_role_code` (`code`),
	KEY `idx_sys_role_deleted_at` (`deleted_at`),
  KEY `idx_sys_role_parent_id` (`parent_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色表';

-- ------------------------------------------------------------
-- 菜单/按钮表（多级；type=3 按钮必须绑定 api_path）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_menu`;
CREATE TABLE `sys_menu` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `deleted_at` DATETIME(3) NULL DEFAULT NULL,
  `name`       VARCHAR(32)  NOT NULL COMMENT '名称',
  `type`       TINYINT NOT NULL DEFAULT 2 COMMENT '1目录 2菜单 3按钮',
  `parent_id`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '上级ID，0=顶级',
  `path`       VARCHAR(128) NOT NULL DEFAULT '' COMMENT '前端路由地址',
  `api_path`   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '后端接口，多个逗号分隔，如 POST:/admin/user/add',
  `icon`       VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '图标',
  `sort`       INT NOT NULL DEFAULT 0 COMMENT '排序',
  `status`     TINYINT NOT NULL DEFAULT 1 COMMENT '1显示 2隐藏',
  `remark`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
  PRIMARY KEY (`id`),
  KEY `idx_sys_menu_deleted_at` (`deleted_at`),
  KEY `idx_sys_menu_parent_id` (`parent_id`),
  KEY `idx_sys_menu_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='菜单/按钮表';

-- ------------------------------------------------------------
-- 登录流水表（login_id 写入 JWT 并缓存于 Redis，用于主动过期/踢人）
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_user_login`;
CREATE TABLE `sys_user_login` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `login_id`   VARCHAR(64)  NOT NULL COMMENT '登录唯一标识(uuid)',
  `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `username`   VARCHAR(32)  NULL COMMENT '登录名',
  `client`     VARCHAR(16)  NULL DEFAULT '' COMMENT 'admin/api',
  `login_ip`   VARCHAR(64)  NULL DEFAULT '' COMMENT '登录IP',
  `user_agent` VARCHAR(255) NULL DEFAULT '' COMMENT 'UA',
  `login_at`   DATETIME(3)  NULL COMMENT '登录时间',
  `logout_at`  DATETIME(3)  NULL COMMENT '退出时间',
  `status`     TINYINT NOT NULL DEFAULT 1 COMMENT '1在线 2已退出 3被踢下线',
  PRIMARY KEY (`id`),
  UNIQUE KEY `login_id` (`login_id`),
  KEY `idx_sys_user_login_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='登录流水表';

-- ------------------------------------------------------------
-- 关联表
-- ------------------------------------------------------------
-- ------------------------------------------------------------
-- 文件存储渠道配置表
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_storage_config`;
CREATE TABLE `sys_storage_config` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at` DATETIME(3) NULL DEFAULT NULL,
  `updated_at` DATETIME(3) NULL DEFAULT NULL,
  `deleted_at` DATETIME(3) NULL DEFAULT NULL,
  `name`       VARCHAR(64) NOT NULL COMMENT '渠道名称',
  `channel`    VARCHAR(32) NOT NULL COMMENT '渠道类型: local/alioss/cos/qiniu/minio',
  `params`     TEXT COMMENT '渠道参数JSON',
  `is_default` TINYINT NOT NULL DEFAULT 0 COMMENT '1默认渠道',
  `status`     TINYINT NOT NULL DEFAULT 1 COMMENT '1启用 2禁用',
  `sort`       INT NOT NULL DEFAULT 0 COMMENT '排序',
  `remark`     VARCHAR(255) NOT NULL DEFAULT '' COMMENT '备注',
	`default_slot` TINYINT GENERATED ALWAYS AS (
		CASE WHEN `deleted_at` IS NULL AND `is_default` = 1 THEN 1 ELSE NULL END
	) STORED COMMENT '保证全局至多一个有效默认渠道',
	PRIMARY KEY (`id`),
	UNIQUE KEY `uk_sys_storage_default_slot` (`default_slot`),
	KEY `idx_sys_storage_config_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='文件存储渠道配置表';

-- ------------------------------------------------------------
-- 上传文件记录表
-- ------------------------------------------------------------
DROP TABLE IF EXISTS `sys_upload_file`;
CREATE TABLE `sys_upload_file` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at`    DATETIME(3) NULL DEFAULT NULL,
  `updated_at`    DATETIME(3) NULL DEFAULT NULL,
  `file_name`     VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `hash_name`     VARCHAR(128) NOT NULL COMMENT '平台hash名(md5+扩展名)',
  `size`          BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `mime_type`     VARCHAR(128) NOT NULL DEFAULT '' COMMENT '文件类型(MIME)',
  `ext`           VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件扩展名',
  `client`        VARCHAR(16) NOT NULL DEFAULT '' COMMENT '来源: admin/api',
  `relative_path` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件相对路径(云端为对象Key)',
  `absolute_path` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '上传时完整访问地址',
  `uploader_id`   BIGINT UNSIGNED NULL COMMENT '上传人ID(可空)',
  PRIMARY KEY (`id`),
  KEY `idx_sys_upload_file_hash_name` (`hash_name`),
  KEY `idx_sys_upload_file_uploader_id` (`uploader_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='上传文件记录表';

-- 操作日志
DROP TABLE IF EXISTS `sys_operation_log`;
CREATE TABLE `sys_operation_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL,
  `user_id` BIGINT UNSIGNED NULL, `username` VARCHAR(32) NOT NULL DEFAULT '', `method` VARCHAR(16) NOT NULL DEFAULT '',
  `path` VARCHAR(128) NOT NULL DEFAULT '', `request_params` MEDIUMTEXT, `response_params` MEDIUMTEXT,
  `ip` VARCHAR(64) NOT NULL DEFAULT '', `user_agent` VARCHAR(255) NOT NULL DEFAULT '', `code` INT NOT NULL DEFAULT 0,
  `cost_ms` BIGINT NOT NULL DEFAULT 0, `client` VARCHAR(16) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  KEY `idx_sys_operation_log_user_id` (`user_id`), KEY `idx_sys_operation_log_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户操作日志表';

-- 短信开发配置、签名、模板与发送记录
DROP TABLE IF EXISTS `sys_sms_config`;
CREATE TABLE `sys_sms_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `name` VARCHAR(64) NOT NULL, `provider` TINYINT NOT NULL COMMENT '1阿里云 2腾讯云',
  `access_key_id` VARCHAR(128) NOT NULL, `access_key_secret` VARCHAR(255) NOT NULL,
  `endpoint` VARCHAR(255) NOT NULL DEFAULT '', `status` TINYINT NOT NULL DEFAULT 1, `remark` VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`), KEY `idx_sys_sms_config_deleted_at` (`deleted_at`), KEY `idx_sys_sms_config_provider` (`provider`), KEY `idx_sys_sms_config_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信开发配置';

DROP TABLE IF EXISTS `sys_sms_signature`;
CREATE TABLE `sys_sms_signature` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `config_id` BIGINT UNSIGNED NOT NULL, `name` VARCHAR(64) NOT NULL, `sign_code` VARCHAR(128) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 1, `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  KEY `idx_sys_sms_signature_deleted_at` (`deleted_at`), KEY `idx_sys_sms_signature_config_id` (`config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信签名';

DROP TABLE IF EXISTS `sys_sms_template`;
CREATE TABLE `sys_sms_template` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `config_id` BIGINT UNSIGNED NOT NULL, `name` VARCHAR(64) NOT NULL, `template_code` VARCHAR(128) NOT NULL,
  `type` TINYINT NOT NULL COMMENT '1验证码 2通知 3营销', `content` VARCHAR(500) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 1, `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  KEY `idx_sys_sms_template_deleted_at` (`deleted_at`), KEY `idx_sys_sms_template_config_id` (`config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信模板';

DROP TABLE IF EXISTS `sys_sms_send_log`;
CREATE TABLE `sys_sms_send_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `config_id` BIGINT UNSIGNED NOT NULL,
  `signature_id` BIGINT UNSIGNED NOT NULL DEFAULT 0, `template_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `mobile` VARCHAR(32) NOT NULL, `content` VARCHAR(1000) NOT NULL DEFAULT '', `status` TINYINT NOT NULL COMMENT '1待发送 2成功 3失败',
  `provider_message_id` VARCHAR(128) NOT NULL DEFAULT '', `error_message` VARCHAR(500) NOT NULL DEFAULT '', `sent_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`), KEY `idx_sys_sms_send_log_config_id` (`config_id`), KEY `idx_sys_sms_send_log_mobile` (`mobile`), KEY `idx_sys_sms_send_log_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信发送记录';

-- 微信配置
DROP TABLE IF EXISTS `sys_wechat_config`;
CREATE TABLE `sys_wechat_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `name` VARCHAR(64) NOT NULL, `type` TINYINT NOT NULL COMMENT '1公众号 2开放平台 3小程序', `app_id` VARCHAR(128) NOT NULL,
  `app_secret` VARCHAR(255) NOT NULL, `token` VARCHAR(255) NOT NULL DEFAULT '', `aes_key` VARCHAR(255) NOT NULL DEFAULT '',
  `status` TINYINT NOT NULL DEFAULT 1, `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_wechat_config_app_id` (`app_id`), KEY `idx_sys_wechat_config_deleted_at` (`deleted_at`), KEY `idx_sys_wechat_config_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='微信应用配置';

-- 支付配置（同一渠道允许多条）
DROP TABLE IF EXISTS `sys_payment_config`;
CREATE TABLE `sys_payment_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `name` VARCHAR(64) NOT NULL, `channel` TINYINT NOT NULL COMMENT '1微信支付 2支付宝', `app_id` VARCHAR(128) NOT NULL,
  `merchant_id` VARCHAR(128) NOT NULL DEFAULT '', `private_key` TEXT, `public_key` TEXT, `api_v3_key` VARCHAR(255) NOT NULL DEFAULT '',
  `cert_serial_no` VARCHAR(128) NOT NULL DEFAULT '', `notify_url` VARCHAR(500) NOT NULL, `status` TINYINT NOT NULL DEFAULT 1,
  `sort` INT NOT NULL DEFAULT 0, `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  KEY `idx_sys_payment_config_deleted_at` (`deleted_at`), KEY `idx_sys_payment_config_channel` (`channel`), KEY `idx_sys_payment_config_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='支付渠道配置';

DROP TABLE IF EXISTS `sys_user_role`;
CREATE TABLE `sys_user_role` (
  `user_id` BIGINT UNSIGNED NOT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`user_id`, `role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户-角色关联';

DROP TABLE IF EXISTS `sys_role_menu`;
CREATE TABLE `sys_role_menu` (
  `role_id` BIGINT UNSIGNED NOT NULL,
  `menu_id` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`role_id`, `menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='角色-菜单关联';

-- =============================================================
-- 初始数据
-- =============================================================

-- 部门：公司 > 技术部(前端组/后端组) / 运营部
INSERT INTO `sys_dept` (`id`, `name`, `parent_id`, `sort`, `leader`, `remark`, `created_at`, `updated_at`) VALUES
(1, '总公司',   0, 1, '张总',   '', NOW(), NOW()),
(2, '技术部',   1, 1, '李工',   '', NOW(), NOW()),
(3, '运营部',   1, 2, '王运营', '', NOW(), NOW()),
(4, '前端组',   2, 1, '', '', NOW(), NOW()),
(5, '后端组',   2, 2, '', '', NOW(), NOW());

-- 用户：admin 为超级管理员；zhangsan 为普通用户
-- 密码均为 123456（bcrypt）
INSERT INTO `sys_user` (`id`, `username`, `password`, `nickname`, `mobile`, `email`, `dept_id`, `status`, `is_super`, `created_at`, `updated_at`) VALUES
(1, 'admin',    '$2a$10$xBQeertsN46doGwzCyVFx.sJ51JOtSvntXzdv5PMmucSXWZEXfxoe', '超级管理员', '13800000001', 'admin@example.com', 5, 1, 1, NOW(), NOW()),
(2, 'zhangsan', '$2a$10$xBQeertsN46doGwzCyVFx.sJ51JOtSvntXzdv5PMmucSXWZEXfxoe', '张三',       '13800000002', 'zhangsan@example.com', 3, 1, 0, NOW(), NOW());

-- 角色：总经理 > 系统管理员 / 运营管理 > 运营专员
INSERT INTO `sys_role` (`id`, `name`, `code`, `parent_id`, `sort`, `status`, `remark`, `created_at`, `updated_at`) VALUES
(1, '总经理',     'ceo',        0, 1, 1, '顶级角色', NOW(), NOW()),
(2, '系统管理员', 'sys_admin',  1, 1, 1, '负责系统权限配置', NOW(), NOW()),
(3, '运营管理',   'op_manager', 1, 2, 1, '运营负责人', NOW(), NOW()),
(4, '运营专员',   'op_staff',   3, 1, 1, '普通运营', NOW(), NOW());

-- 菜单/按钮
-- 系统总览（页面，带汇总统计接口）
INSERT INTO `sys_menu` (`id`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(1, '系统总览', 2, 0, '/dashboard', 'GET:/admin/dashboard/overview', 'Odometer', 1, 1, NOW(), NOW());

-- 系统权限（目录） > 部门/菜单/角色/人员管理（页面 + 操作按钮）
INSERT INTO `sys_menu` (`id`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(2, '系统权限', 1, 0, '/system', '', 'Setting', 2, 1, NOW(), NOW()),

(3, '部门管理', 2, 2, '/system/dept', 'GET:/admin/dept/tree', 'OfficeBuilding', 1, 1, NOW(), NOW()),
(4, '部门新增', 3, 3, '', 'POST:/admin/dept/add', '', 1, 1, NOW(), NOW()),
(5, '部门修改', 3, 3, '', 'POST:/admin/dept/update', '', 2, 1, NOW(), NOW()),
(6, '部门删除', 3, 3, '', 'POST:/admin/dept/delete', '', 3, 1, NOW(), NOW()),

(7, '菜单管理', 2, 2, '/system/menu', 'GET:/admin/menu/list,GET:/admin/menu/tree', 'Menu', 2, 1, NOW(), NOW()),
(8, '菜单新增', 3, 7, '', 'POST:/admin/menu/add', '', 1, 1, NOW(), NOW()),
(9, '菜单修改', 3, 7, '', 'POST:/admin/menu/update', '', 2, 1, NOW(), NOW()),
(10,'菜单删除', 3, 7, '', 'POST:/admin/menu/delete', '', 3, 1, NOW(), NOW()),

(11,'角色管理', 2, 2, '/system/role', 'GET:/admin/role/list,GET:/admin/role/tree', 'UserFilled', 3, 1, NOW(), NOW()),
(12,'角色新增', 3, 11, '', 'POST:/admin/role/add', '', 1, 1, NOW(), NOW()),
(13,'角色修改', 3, 11, '', 'POST:/admin/role/update', '', 2, 1, NOW(), NOW()),
(14,'角色删除', 3, 11, '', 'POST:/admin/role/delete', '', 3, 1, NOW(), NOW()),
(15,'角色分配权限', 3, 11, '', 'POST:/admin/role/assign_menus', '', 4, 1, NOW(), NOW()),

(16,'人员管理', 2, 2, '/system/user', 'GET:/admin/user/list', 'User', 4, 1, NOW(), NOW()),
(17,'用户新增', 3, 16, '', 'POST:/admin/user/add', '', 1, 1, NOW(), NOW()),
(18,'用户修改', 3, 16, '', 'POST:/admin/user/update', '', 2, 1, NOW(), NOW()),
(19,'用户删除', 3, 16, '', 'POST:/admin/user/delete', '', 3, 1, NOW(), NOW()),
(20,'重置密码', 3, 16, '', 'POST:/admin/user/reset_password', '', 4, 1, NOW(), NOW()),
(21,'踢用户下线', 3, 16, '', 'POST:/admin/user/kick', '', 5, 1, NOW(), NOW());

-- 渠道配置：存储、短信、微信、支付为同级页面
INSERT INTO `sys_menu` (`id`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(22,'系统配置',1,0,'/config','','Tools',3,1,NOW(),NOW()),
(23,'渠道配置',1,22,'/config/channel','','Connection',1,1,NOW(),NOW()),
(24,'存储配置',2,23,'/config/channel/storage','GET:/admin/storage/list','Box',1,1,NOW(),NOW()),
(25,'存储新增',3,24,'','POST:/admin/storage/add','',1,1,NOW(),NOW()),
(26,'存储修改',3,24,'','POST:/admin/storage/update','',2,1,NOW(),NOW()),
(27,'存储删除',3,24,'','POST:/admin/storage/delete','',3,1,NOW(),NOW()),
(28,'存储设为默认',3,24,'','POST:/admin/storage/set_default','',4,1,NOW(),NOW()),
(29,'系统维护',1,0,'/maintain','','Hammer',4,1,NOW(),NOW()),
(30,'日志维护',1,29,'/maintain/log','','Document',1,1,NOW(),NOW()),
(31,'操作日志',2,30,'/maintain/log/operation','GET:/admin/log/operation/list','List',1,1,NOW(),NOW()),
(32,'短信配置',2,23,'/config/channel/sms','GET:/admin/sms/config/list,GET:/admin/sms/signature/list,GET:/admin/sms/template/list,GET:/admin/sms/log/list','ChatDotRound',2,1,NOW(),NOW()),
(33,'短信开发配置保存',3,32,'','POST:/admin/sms/config/save','',1,1,NOW(),NOW()),
(34,'短信开发配置删除',3,32,'','POST:/admin/sms/config/delete','',2,1,NOW(),NOW()),
(35,'短信签名保存',3,32,'','POST:/admin/sms/signature/save','',3,1,NOW(),NOW()),
(36,'短信签名删除',3,32,'','POST:/admin/sms/signature/delete','',4,1,NOW(),NOW()),
(37,'短信模板保存',3,32,'','POST:/admin/sms/template/save','',5,1,NOW(),NOW()),
(38,'短信模板删除',3,32,'','POST:/admin/sms/template/delete','',6,1,NOW(),NOW()),
(39,'微信配置',2,23,'/config/channel/wechat','GET:/admin/wechat/config/list','ChatLineRound',3,1,NOW(),NOW()),
(40,'微信配置保存',3,39,'','POST:/admin/wechat/config/save','',1,1,NOW(),NOW()),
(41,'微信配置删除',3,39,'','POST:/admin/wechat/config/delete','',2,1,NOW(),NOW()),
(42,'支付配置',2,23,'/config/channel/payment','GET:/admin/payment/config/list','Wallet',4,1,NOW(),NOW()),
(43,'支付配置保存',3,42,'','POST:/admin/payment/config/save','',1,1,NOW(),NOW()),
(44,'支付配置删除',3,42,'','POST:/admin/payment/config/delete','',2,1,NOW(),NOW());

INSERT INTO `sys_storage_config` (`name`, `channel`, `params`, `is_default`, `status`, `sort`, `remark`, `created_at`, `updated_at`) VALUES
('本地存储', 'local', '{"root_path":"./uploads"}', 1, 1, 1, '默认本地文件存储', NOW(), NOW());

-- 角色菜单绑定
-- 系统管理员(id=2)：全部菜单与按钮
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`)
SELECT 2, `id` FROM `sys_menu`;

-- 运营专员(id=4)：仅系统总览
INSERT INTO `sys_role_menu` (`role_id`, `menu_id`) VALUES (4, 1);

-- 用户角色绑定
INSERT INTO `sys_user_role` (`user_id`, `role_id`) VALUES (2, 4);
