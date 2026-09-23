-- =============================================================
-- 增量升级脚本：存储渠道配置功能（已初始化过的库执行本文件即可）
-- 在与 schema.sql 相同的业务库中执行，全新安装无需执行本文件。
-- =============================================================

-- 1. 存储渠道配置表
CREATE TABLE IF NOT EXISTS `sys_storage_config` (
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
  PRIMARY KEY (`id`),
  KEY `idx_sys_storage_config_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='文件存储渠道配置表';

-- 2. 菜单：系统配置 > 渠道配置 > 存储配置（+ 操作按钮，ID 基于初始种子最大 21）
INSERT IGNORE INTO `sys_menu` (`id`, `name`, `type`, `parent_id`, `path`, `api_path`, `icon`, `sort`, `status`, `created_at`, `updated_at`) VALUES
(22, '系统配置', 1, 0,  '/config',                '',                    'Tools',      2, 1, NOW(), NOW()),
(23, '渠道配置', 1, 22, '/config/channel',        '',                    'Connection', 1, 1, NOW(), NOW()),
(24, '存储配置', 2, 23, '/config/channel/storage','GET:/admin/storage/list', 'Box',   1, 1, NOW(), NOW()),
(25, '存储新增', 3, 24, '', 'POST:/admin/storage/add',         '', 1, 1, NOW(), NOW()),
(26, '存储修改', 3, 24, '', 'POST:/admin/storage/update',      '', 2, 1, NOW(), NOW()),
(27, '存储删除', 3, 24, '', 'POST:/admin/storage/delete',      '', 3, 1, NOW(), NOW()),
(28, '存储设为默认', 3, 24, '', 'POST:/admin/storage/set_default', '', 4, 1, NOW(), NOW());

-- 系统权限下移一位，让系统配置排在其上方
UPDATE `sys_menu` SET `sort` = 3 WHERE `name` = '系统权限' AND `parent_id` = 0;

-- 3. 本地存储默认渠道
INSERT INTO `sys_storage_config` (`name`, `channel`, `params`, `is_default`, `status`, `sort`, `remark`, `created_at`, `updated_at`)
SELECT '本地存储', 'local', '{"root_path":"./uploads"}', 1, 1, 1, '默认本地文件存储', NOW(), NOW()
WHERE NOT EXISTS (SELECT 1 FROM `sys_storage_config`);

-- 4. 系统管理员角色（role_id=2）绑定新菜单
INSERT IGNORE INTO `sys_role_menu` (`role_id`, `menu_id`)
SELECT 2, `id` FROM `sys_menu` WHERE `id` >= 22;
