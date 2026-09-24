-- 插件菜单支持管理员自定义父级；脚本可重复执行。
SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

SET @plugin_parent_source_exists := (
  SELECT COUNT(*) FROM `information_schema`.`COLUMNS`
  WHERE `TABLE_SCHEMA` = DATABASE()
    AND `TABLE_NAME` = 'sys_plugin_menu'
    AND `COLUMN_NAME` = 'parent_source'
);
SET @plugin_parent_source_sql := IF(
  @plugin_parent_source_exists = 0,
  'ALTER TABLE `sys_plugin_menu` ADD COLUMN `parent_source` tinyint unsigned NOT NULL DEFAULT ''1'' COMMENT ''父级来源，1插件默认，2管理员自定义'' AFTER `parent_key`',
  'SELECT 1'
);
PREPARE plugin_parent_source_stmt FROM @plugin_parent_source_sql;
EXECUTE plugin_parent_source_stmt;
DEALLOCATE PREPARE plugin_parent_source_stmt;

-- 兼容功能上线前已经由管理员移动过的插件目录和页面。
UPDATE `sys_plugin_menu` AS `plugin_menu`
JOIN `sys_menu` AS `menu` ON `menu`.`id` = `plugin_menu`.`menu_id`
LEFT JOIN `sys_plugin_menu` AS `parent_mapping`
  ON `parent_mapping`.`plugin_id` = `plugin_menu`.`plugin_id`
  AND `parent_mapping`.`menu_key` = `plugin_menu`.`parent_key`
SET `plugin_menu`.`parent_source` = 2
WHERE `menu`.`type` <> 3
  AND `menu`.`parent_id` <> CASE
    WHEN `plugin_menu`.`parent_key` = '' THEN 0
    ELSE COALESCE(`parent_mapping`.`menu_id`, 0)
  END;
