-- v0.0.2 插件化基础设施升级脚本
-- 执行前请备份数据库；本脚本只新增核心表，不修改现有业务数据。

SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `sys_plugin` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间',
  `plugin_id` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '插件唯一标识',
  `name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL COMMENT '插件名称',
  `version` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '已安装插件版本',
  `logo` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '插件Logo相对路径',
  `author` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '插件作者',
  `homepage` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '插件主页地址',
  `description` varchar(2000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '插件描述',
  `status` tinyint unsigned NOT NULL DEFAULT '2' COMMENT '状态，1启用，2停用',
  `manifest` mediumtext COLLATE utf8mb4_general_ci COMMENT '安装时插件清单JSON快照',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_sys_plugin_plugin_id` (`plugin_id`),
  KEY `idx_sys_plugin_deleted_at` (`deleted_at`),
  KEY `idx_sys_plugin_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件信息表';

-- 兼容已经执行过早期 v0.0.2 脚本的数据库，按需补充插件展示字段。
SET @plugin_logo_sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_plugin' AND COLUMN_NAME = 'logo') = 0,
  'ALTER TABLE `sys_plugin` ADD COLUMN `logo` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '''' COMMENT ''插件Logo相对路径'' AFTER `version`',
  'SELECT 1'
);
PREPARE plugin_logo_stmt FROM @plugin_logo_sql;
EXECUTE plugin_logo_stmt;
DEALLOCATE PREPARE plugin_logo_stmt;

SET @plugin_author_sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_plugin' AND COLUMN_NAME = 'author') = 0,
  'ALTER TABLE `sys_plugin` ADD COLUMN `author` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '''' COMMENT ''插件作者'' AFTER `logo`',
  'SELECT 1'
);
PREPARE plugin_author_stmt FROM @plugin_author_sql;
EXECUTE plugin_author_stmt;
DEALLOCATE PREPARE plugin_author_stmt;

SET @plugin_homepage_sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_plugin' AND COLUMN_NAME = 'homepage') = 0,
  'ALTER TABLE `sys_plugin` ADD COLUMN `homepage` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '''' COMMENT ''插件主页地址'' AFTER `author`',
  'SELECT 1'
);
PREPARE plugin_homepage_stmt FROM @plugin_homepage_sql;
EXECUTE plugin_homepage_stmt;
DEALLOCATE PREPARE plugin_homepage_stmt;

SET @plugin_description_sql := IF(
  (SELECT COUNT(*) FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'sys_plugin' AND COLUMN_NAME = 'description') = 0,
  'ALTER TABLE `sys_plugin` ADD COLUMN `description` varchar(2000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '''' COMMENT ''插件描述'' AFTER `homepage`',
  'SELECT 1'
);
PREPARE plugin_description_stmt FROM @plugin_description_sql;
EXECUTE plugin_description_stmt;
DEALLOCATE PREPARE plugin_description_stmt;

CREATE TABLE IF NOT EXISTS `sys_plugin_migration` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `plugin_id` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '插件唯一标识',
  `version` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '迁移版本',
  `checksum` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '迁移内容SHA256摘要',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '执行状态，1成功，2失败',
  `execution_ms` bigint NOT NULL DEFAULT '0' COMMENT '执行耗时毫秒',
  `error_message` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '失败原因',
  `executed_at` datetime(3) NOT NULL COMMENT '执行时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plugin_migration` (`plugin_id`, `version`),
  KEY `idx_plugin_migration_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件数据库迁移记录表';

CREATE TABLE IF NOT EXISTS `sys_plugin_menu` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `plugin_id` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '插件唯一标识',
  `menu_key` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '插件内菜单业务键',
  `menu_id` bigint unsigned NOT NULL COMMENT '系统菜单ID',
  `parent_key` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '插件内父级菜单业务键',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plugin_menu_key` (`plugin_id`, `menu_key`),
  UNIQUE KEY `uk_plugin_menu_id` (`menu_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件菜单业务键映射表';

CREATE TABLE IF NOT EXISTS `sys_plugin_install_log` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `created_at` datetime(3) NOT NULL COMMENT '创建时间',
  `plugin_id` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '插件唯一标识',
  `version` varchar(32) COLLATE utf8mb4_general_ci NOT NULL COMMENT '目标版本',
  `action` tinyint unsigned NOT NULL COMMENT '操作类型，1安装，2升级',
  `status` tinyint unsigned NOT NULL COMMENT '执行状态，1成功，2失败',
  `package_hash` varchar(64) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '发行包SHA256',
  `error_message` varchar(1000) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '失败原因',
  PRIMARY KEY (`id`),
  KEY `idx_plugin_install_log_plugin_id` (`plugin_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='插件安装升级日志表';

-- 插件管理菜单。使用路径和接口权限判断是否存在，脚本可重复执行。
SET @plugin_menu_parent_id := (
  SELECT `id` FROM `sys_menu`
  WHERE `path` = '/maintain' AND `deleted_at` IS NULL
  ORDER BY `id` ASC LIMIT 1
);

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT
  NOW(3), NOW(3), NULL, '插件管理', 2, @plugin_menu_parent_id, '/maintain/plugin',
  'GET:/admin/plugin/list,GET:/admin/plugin/detail', 'Grid', 2, 1, '查看插件安装、编译及迁移状态'
WHERE @plugin_menu_parent_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `path` = '/maintain/plugin' AND `deleted_at` IS NULL
  );

SET @plugin_menu_id := (
  SELECT `id` FROM `sys_menu`
  WHERE `path` = '/maintain/plugin' AND `deleted_at` IS NULL
  ORDER BY `id` ASC LIMIT 1
);

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT
  NOW(3), NOW(3), NULL, '插件启停', 3, @plugin_menu_id, '',
  'POST:/admin/plugin/status', '', 1, 1, '启用或停用插件，重启服务后生效'
WHERE @plugin_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `api_path` = 'POST:/admin/plugin/status' AND `deleted_at` IS NULL
  );

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT
  NOW(3), NOW(3), NULL, '修改插件信息', 3, @plugin_menu_id, '',
  'POST:/admin/plugin/info', '', 2, 1, '维护插件作者、主页地址和描述'
WHERE @plugin_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `api_path` = 'POST:/admin/plugin/info' AND `deleted_at` IS NULL
  );
