-- 并发一致性升级脚本（已有数据库手动执行）
-- 执行前请先确认 sys_role.code 没有重复值，并确保最多只有一个未删除的默认存储渠道。

ALTER TABLE `sys_role`
  ADD UNIQUE KEY `uk_sys_role_code` (`code`);

ALTER TABLE `sys_storage_config`
  ADD COLUMN `default_slot` TINYINT GENERATED ALWAYS AS (
    CASE WHEN `deleted_at` IS NULL AND `is_default` = 1 THEN 1 ELSE NULL END
  ) STORED COMMENT '保证全局至多一个有效默认渠道',
  ADD UNIQUE KEY `uk_sys_storage_default_slot` (`default_slot`);
