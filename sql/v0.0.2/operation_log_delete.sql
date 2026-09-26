-- 操作日志批量删除与全量清空权限；按接口路径幂等写入，不依赖固定菜单ID。
SET NAMES utf8mb4 COLLATE utf8mb4_general_ci;

SET @operation_log_menu_id := (
  SELECT `id` FROM `sys_menu`
  WHERE `path` = '/maintain/log/operation'
    AND `type` = 2
    AND `deleted_at` IS NULL
  ORDER BY `id` ASC LIMIT 1
);

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT
  NOW(3), NOW(3), NULL, '操作日志批量删除', 3, @operation_log_menu_id, '',
  'POST:/admin/log/operation/delete', '', 1, 1, '物理删除勾选的操作日志'
WHERE @operation_log_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `api_path` = 'POST:/admin/log/operation/delete'
      AND `deleted_at` IS NULL
  );

INSERT INTO `sys_menu`
  (`created_at`, `updated_at`, `deleted_at`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `remark`)
SELECT
  NOW(3), NOW(3), NULL, '操作日志全量清空', 3, @operation_log_menu_id, '',
  'POST:/admin/log/operation/clear', '', 2, 1, '物理删除全部操作日志'
WHERE @operation_log_menu_id IS NOT NULL
  AND NOT EXISTS (
    SELECT 1 FROM `sys_menu`
    WHERE `api_path` = 'POST:/admin/log/operation/clear'
      AND `deleted_at` IS NULL
  );
