-- v0.0.3 短信默认渠道升级脚本。
-- 作用：允许多个短信渠道启用，但数据库层保证未删除配置至多一个默认渠道。
-- 兼容 MySQL 5.7/8.0；可重复执行，不指定任何业务数据主键。

SET @schema_name = DATABASE();

SELECT COUNT(*) INTO @has_is_default
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = @schema_name
  AND TABLE_NAME = 'sys_sms_config'
  AND COLUMN_NAME = 'is_default';

SET @ddl = IF(
  @has_is_default = 0,
  'ALTER TABLE `sys_sms_config` ADD COLUMN `is_default` tinyint NOT NULL DEFAULT 0 COMMENT ''是否默认渠道，0否1是'' AFTER `endpoint`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 禁用配置不能作为默认渠道；异常历史数据只保留 ID 最小的启用默认渠道。
UPDATE `sys_sms_config`
SET `is_default` = 0
WHERE `is_default` = 1 AND `status` <> 1;

UPDATE `sys_sms_config`
SET `is_default` = 0
WHERE `is_default` = 1
  AND `id` <> (
    SELECT `selected`.`id`
    FROM (
      SELECT MIN(`id`) AS `id`
      FROM `sys_sms_config`
      WHERE `deleted_at` IS NULL AND `status` = 1 AND `is_default` = 1
    ) AS `selected`
  );

-- 升级旧数据时如果还没有默认渠道，沿用旧逻辑选择 ID 最小的启用渠道。
UPDATE `sys_sms_config`
SET `is_default` = 1
WHERE `id` = (
  SELECT `selected`.`id`
  FROM (
    SELECT MIN(`id`) AS `id`
    FROM `sys_sms_config`
    WHERE `deleted_at` IS NULL AND `status` = 1
  ) AS `selected`
)
AND NOT EXISTS (
  SELECT 1
  FROM (
    SELECT `id`
    FROM `sys_sms_config`
    WHERE `deleted_at` IS NULL AND `is_default` = 1
    LIMIT 1
  ) AS `current_default`
);

SELECT COUNT(*) INTO @has_default_slot
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = @schema_name
  AND TABLE_NAME = 'sys_sms_config'
  AND COLUMN_NAME = 'default_slot';

SET @ddl = IF(
  @has_default_slot = 0,
  'ALTER TABLE `sys_sms_config` ADD COLUMN `default_slot` tinyint GENERATED ALWAYS AS (CASE WHEN (`deleted_at` IS NULL AND `is_default` = 1) THEN 1 ELSE NULL END) STORED COMMENT ''保证全局至多一个有效默认短信渠道''',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @has_default_unique
FROM information_schema.STATISTICS
WHERE TABLE_SCHEMA = @schema_name
  AND TABLE_NAME = 'sys_sms_config'
  AND INDEX_NAME = 'uk_sys_sms_config_default_slot';

SET @ddl = IF(
  @has_default_unique = 0,
  'ALTER TABLE `sys_sms_config` ADD UNIQUE KEY `uk_sys_sms_config_default_slot` (`default_slot`)',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
