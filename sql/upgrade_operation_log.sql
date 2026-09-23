-- =============================================================
-- 增量升级脚本：系统维护/操作日志（已初始化过的库执行本文件即可）
-- 全新安装无需执行（schema.sql 已包含）。
-- =============================================================

-- 1. 操作日志表
CREATE TABLE IF NOT EXISTS `sys_operation_log` (
  `id`              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at`      DATETIME(3) NULL DEFAULT NULL COMMENT '操作时间',
  `updated_at`      DATETIME(3) NULL DEFAULT NULL,
  `user_id`         BIGINT UNSIGNED NULL COMMENT '操作人ID',
  `username`        VARCHAR(32) NOT NULL DEFAULT '' COMMENT '操作人账号',
  `method`          VARCHAR(16) NOT NULL DEFAULT '' COMMENT '请求方式',
  `path`            VARCHAR(128) NOT NULL DEFAULT '' COMMENT '请求路由',
  `request_params`  MEDIUMTEXT COMMENT '请求参数(敏感字段脱敏)',
  `response_params` MEDIUMTEXT COMMENT '接口响应内容(不截断)',
  `ip`              VARCHAR(64) NOT NULL DEFAULT '' COMMENT '操作IP',
  `user_agent`      VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'UA',
  `code`            INT NOT NULL DEFAULT 0 COMMENT '业务响应码，0=成功',
  `cost_ms`         BIGINT NOT NULL DEFAULT 0 COMMENT '耗时(毫秒)',
  `client`          VARCHAR(16) NOT NULL DEFAULT '' COMMENT '来源: admin/api',
  PRIMARY KEY (`id`),
  KEY `idx_sys_operation_log_user_id` (`user_id`),
  KEY `idx_sys_operation_log_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户操作日志表';

-- 2. 菜单：系统维护 > 日志维护 > 操作日志（ID 基于存储功能种子最大 28）
INSERT IGNORE INTO `sys_menu` (`id`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(29, '系统维护', 1, 0,  '/maintain',               '',                          'Hammer',   4, 1, NOW(), NOW()),
(30, '日志维护', 1, 29, '/maintain/log',           '',                          'Document', 1, 1, NOW(), NOW()),
(31, '操作日志', 2, 30, '/maintain/log/operation', 'GET:/admin/log/operation/list', 'List', 1, 1, NOW(), NOW());

-- 3. 系统管理员角色（role_id=2）绑定新菜单
INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`)
SELECT 2, `id` FROM `sys_menu` WHERE `id` >= 29;

-- 3. 已建过的表升级为 MEDIUMTEXT（不截断）
ALTER TABLE `sys_operation_log` MODIFY COLUMN `request_params` MEDIUMTEXT COMMENT '请求参数(敏感字段脱敏)';
ALTER TABLE `sys_operation_log` MODIFY COLUMN `response_params` MEDIUMTEXT COMMENT '接口响应内容(不截断)';
