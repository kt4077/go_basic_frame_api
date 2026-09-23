-- 平台配置表及菜单增量脚本
-- 执行方式：mysql -uroot -p 数据库名 < sql/platform_config.sql

SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

CREATE TABLE IF NOT EXISTS `sys_platform_config` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `created_at` datetime(3) DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime(3) DEFAULT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间',
  `type` tinyint unsigned NOT NULL COMMENT '平台类型，1管理端，2用户端',
  `logo` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '管理端Logo相对路径',
  `system_name` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '管理端系统名称',
  `default_nickname` varchar(100) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户端默认昵称',
  `default_avatar` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '用户端默认头像相对路径',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_platform_type` (`type`),
  KEY `idx_sys_platform_config_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='平台基础配置表';

INSERT INTO `sys_platform_config`
  (`created_at`, `updated_at`, `type`, `logo`, `system_name`, `default_nickname`, `default_avatar`)
SELECT NOW(3), NOW(3), 1, '', '后台管理系统', '', ''
WHERE NOT EXISTS (SELECT 1 FROM `sys_platform_config` WHERE `type` = 1 AND `deleted_at` IS NULL);

INSERT INTO `sys_platform_config`
  (`created_at`, `updated_at`, `type`, `logo`, `system_name`, `default_nickname`, `default_avatar`)
SELECT NOW(3), NOW(3), 2, '', '', '用户', ''
WHERE NOT EXISTS (SELECT 1 FROM `sys_platform_config` WHERE `type` = 2 AND `deleted_at` IS NULL);

-- 平台配置是“系统配置”下的二级目录。
SET @system_config_id := (
  SELECT `id` FROM `sys_menu`
  WHERE `name` = '系统配置' AND `parent_id` = 0 AND `deleted_at` IS NULL
  ORDER BY `id` LIMIT 1
);

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '平台配置', 1, @system_config_id, '', '', 'SetUp', 10, 1, '管理各端平台基础信息'
WHERE @system_config_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `name` = '平台配置' AND `parent_id` = @system_config_id AND `deleted_at` IS NULL
  );

SET @platform_config_id := (
  SELECT `id` FROM `sys_menu`
  WHERE `name` = '平台配置' AND `parent_id` = @system_config_id AND `deleted_at` IS NULL
  ORDER BY `id` LIMIT 1
);

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '管理端配置', 2, @platform_config_id, '/config/platform/admin', 'GET:/admin/platform/admin/detail', 'Monitor', 1, 1, '管理端Logo与系统名称'
WHERE @platform_config_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `path` = '/config/platform/admin' AND `deleted_at` IS NULL
  );

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '用户端配置', 2, @platform_config_id, '/config/platform/user', 'GET:/admin/platform/user/detail', 'UserFilled', 2, 1, '用户默认昵称与头像'
WHERE @platform_config_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `path` = '/config/platform/user' AND `deleted_at` IS NULL
  );

SET @admin_config_id := (
  SELECT `id` FROM `sys_menu` WHERE `path` = '/config/platform/admin' AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1
);
SET @user_config_id := (
  SELECT `id` FROM `sys_menu` WHERE `path` = '/config/platform/user' AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1
);

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '保存管理端配置', 3, @admin_config_id, '', 'POST:/admin/platform/admin/save', '', 1, 1, '保存管理端平台配置'
WHERE @admin_config_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `api_path` = 'POST:/admin/platform/admin/save' AND `deleted_at` IS NULL
  );

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT NOW(3), NOW(3), '保存用户端配置', 3, @user_config_id, '', 'POST:/admin/platform/user/save', '', 1, 1, '保存用户端平台配置'
WHERE @user_config_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `api_path` = 'POST:/admin/platform/user/save' AND `deleted_at` IS NULL
  );

-- 已拥有“系统配置”的角色继承平台配置及其子菜单权限。
INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`)
SELECT role_source.`role_id`, menu_source.`id`
FROM `sys_role_menu` AS role_source
JOIN `sys_menu` AS menu_source
  ON menu_source.`id` IN (
    @platform_config_id,
    @admin_config_id,
    @user_config_id,
    (SELECT `id` FROM `sys_menu` WHERE `api_path` = 'POST:/admin/platform/admin/save' AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1),
    (SELECT `id` FROM `sys_menu` WHERE `api_path` = 'POST:/admin/platform/user/save' AND `deleted_at` IS NULL ORDER BY `id` LIMIT 1)
  )
WHERE role_source.`menu_id` = @system_config_id;
