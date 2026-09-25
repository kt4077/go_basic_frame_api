-- 用户管理：会员用户表 + 管理端菜单（用户管理 → 系统用户）。脚本可重复执行。
SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `sys_member` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间',
  `nickname` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '昵称',
  `real_name` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '姓名',
  `account` varchar(32) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '登录账号',
  `mobile` varchar(16) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '手机号',
  `password` varchar(128) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '登录密码bcrypt密文',
  `avatar` varchar(512) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '头像相对路径',
  `gender` tinyint NOT NULL DEFAULT '3' COMMENT '性别：1男，2女，3未知',
  `age` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '年龄',
  `birthday` date DEFAULT NULL COMMENT '出生日期',
  `register_ip` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '注册IP',
  `login_ip` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '最近登录IP',
  `registered_at` datetime(3) DEFAULT NULL COMMENT '注册时间',
  `logged_at` datetime(3) DEFAULT NULL COMMENT '最近登录时间',
  `balance` decimal(12,2) NOT NULL DEFAULT '0.00' COMMENT '账户余额',
  `register_source` tinyint NOT NULL DEFAULT '1' COMMENT '注册来源：1微信小程序，2微信公众号，3iOS，4Android',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '账号状态：1启用，2禁用',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_member_mobile` (`mobile`),
  KEY `idx_sys_member_deleted_at` (`deleted_at`),
  KEY `idx_sys_member_status` (`status`),
  KEY `idx_sys_member_account` (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='会员用户表';

-- 菜单通过 path/api_path 业务键判断是否已存在，不指定自增ID；普通角色权限由管理员在角色管理中分配。
INSERT INTO `sys_menu` (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '用户管理', 1, 0, '/user', '', 'Avatar', 5, 1, '用户管理目录'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/user' AND `type` = 1 AND `deleted_at` IS NULL);

SET @member_dir_id := (
  SELECT `id` FROM `sys_menu` WHERE `path` = '/user' AND `type` = 1 AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1
);

INSERT INTO `sys_menu` (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '系统用户', 2, @member_dir_id, '/user/system', 'GET:/admin/member/list', 'User', 1, 1, '系统用户列表'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `path` = '/user/system' AND `type` = 2 AND `deleted_at` IS NULL);

SET @member_page_id := (
  SELECT `id` FROM `sys_menu` WHERE `path` = '/user/system' AND `type` = 2 AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1
);

INSERT INTO `sys_menu` (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '用户禁用', 3, @member_page_id, '', 'POST:/admin/member/set_status', '', 1, 1, '启用/禁用系统用户账号'
FROM DUAL
WHERE NOT EXISTS (SELECT 1 FROM `sys_menu` WHERE `api_path` = 'POST:/admin/member/set_status' AND `type` = 3 AND `deleted_at` IS NULL);
