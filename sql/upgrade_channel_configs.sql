-- 渠道配置增量升级：短信、微信、支付及菜单。全新安装只执行 schema.sql。

CREATE TABLE IF NOT EXISTS `sys_sms_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `name` VARCHAR(64) NOT NULL, `provider` TINYINT NOT NULL, `access_key_id` VARCHAR(128) NOT NULL, `access_key_secret` VARCHAR(255) NOT NULL,
  `endpoint` VARCHAR(255) NOT NULL DEFAULT '', `status` TINYINT NOT NULL DEFAULT 1, `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  KEY `idx_sys_sms_config_deleted_at` (`deleted_at`), KEY `idx_sys_sms_config_provider` (`provider`), KEY `idx_sys_sms_config_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信开发配置';
CREATE TABLE IF NOT EXISTS `sys_sms_signature` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `config_id` BIGINT UNSIGNED NOT NULL, `name` VARCHAR(64) NOT NULL, `sign_code` VARCHAR(128) NOT NULL DEFAULT '', `status` TINYINT NOT NULL DEFAULT 1,
  `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`), KEY `idx_sys_sms_signature_deleted_at` (`deleted_at`), KEY `idx_sys_sms_signature_config_id` (`config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信签名';
CREATE TABLE IF NOT EXISTS `sys_sms_template` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `config_id` BIGINT UNSIGNED NOT NULL, `name` VARCHAR(64) NOT NULL, `template_code` VARCHAR(128) NOT NULL, `type` TINYINT NOT NULL,
  `content` VARCHAR(500) NOT NULL DEFAULT '', `status` TINYINT NOT NULL DEFAULT 1, `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`),
  KEY `idx_sys_sms_template_deleted_at` (`deleted_at`), KEY `idx_sys_sms_template_config_id` (`config_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信模板';
CREATE TABLE IF NOT EXISTS `sys_sms_send_log` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `config_id` BIGINT UNSIGNED NOT NULL,
  `signature_id` BIGINT UNSIGNED NOT NULL DEFAULT 0, `template_id` BIGINT UNSIGNED NOT NULL DEFAULT 0, `mobile` VARCHAR(32) NOT NULL,
  `content` VARCHAR(1000) NOT NULL DEFAULT '', `status` TINYINT NOT NULL, `provider_message_id` VARCHAR(128) NOT NULL DEFAULT '',
  `error_message` VARCHAR(500) NOT NULL DEFAULT '', `sent_at` DATETIME(3) NULL, PRIMARY KEY (`id`), KEY `idx_sys_sms_send_log_config_id` (`config_id`),
  KEY `idx_sys_sms_send_log_mobile` (`mobile`), KEY `idx_sys_sms_send_log_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='短信发送记录';
CREATE TABLE IF NOT EXISTS `sys_wechat_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `name` VARCHAR(64) NOT NULL, `type` TINYINT NOT NULL, `app_id` VARCHAR(128) NOT NULL, `app_secret` VARCHAR(255) NOT NULL,
  `token` VARCHAR(255) NOT NULL DEFAULT '', `aes_key` VARCHAR(255) NOT NULL DEFAULT '', `status` TINYINT NOT NULL DEFAULT 1,
  `remark` VARCHAR(255) NOT NULL DEFAULT '', PRIMARY KEY (`id`), UNIQUE KEY `uk_sys_wechat_config_app_id` (`app_id`),
  KEY `idx_sys_wechat_config_deleted_at` (`deleted_at`), KEY `idx_sys_wechat_config_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='微信应用配置';
CREATE TABLE IF NOT EXISTS `sys_payment_config` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, `created_at` DATETIME(3) NULL, `updated_at` DATETIME(3) NULL, `deleted_at` DATETIME(3) NULL,
  `name` VARCHAR(64) NOT NULL, `channel` TINYINT NOT NULL, `app_id` VARCHAR(128) NOT NULL, `merchant_id` VARCHAR(128) NOT NULL DEFAULT '',
  `private_key` TEXT, `public_key` TEXT, `api_v3_key` VARCHAR(255) NOT NULL DEFAULT '', `cert_serial_no` VARCHAR(128) NOT NULL DEFAULT '',
  `notify_url` VARCHAR(500) NOT NULL, `status` TINYINT NOT NULL DEFAULT 1, `sort` INT NOT NULL DEFAULT 0, `remark` VARCHAR(255) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`), KEY `idx_sys_payment_config_deleted_at` (`deleted_at`), KEY `idx_sys_payment_config_channel` (`channel`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='支付渠道配置';

INSERT IGNORE INTO `sys_menu` (`id`,`name`,`type`,`parent_id`,`path`,`api_path`,`icon`,`sort`,`status`,`created_at`,`updated_at`) VALUES
(22,'系统配置',1,0,'/config','','Tools',3,1,NOW(),NOW()),
(23,'渠道配置',1,22,'/config/channel','','Connection',1,1,NOW(),NOW()),
(32,'短信配置',2,23,'/config/channel/sms','GET:/admin/sms/config/list,GET:/admin/sms/signature/list,GET:/admin/sms/template/list,GET:/admin/sms/log/list','ChatDotRound',2,1,NOW(),NOW()),
(33,'短信开发配置保存',3,32,'','POST:/admin/sms/config/save','',1,1,NOW(),NOW()),(34,'短信开发配置删除',3,32,'','POST:/admin/sms/config/delete','',2,1,NOW(),NOW()),
(35,'短信签名保存',3,32,'','POST:/admin/sms/signature/save','',3,1,NOW(),NOW()),(36,'短信签名删除',3,32,'','POST:/admin/sms/signature/delete','',4,1,NOW(),NOW()),
(37,'短信模板保存',3,32,'','POST:/admin/sms/template/save','',5,1,NOW(),NOW()),(38,'短信模板删除',3,32,'','POST:/admin/sms/template/delete','',6,1,NOW(),NOW()),
(39,'微信配置',2,23,'/config/channel/wechat','GET:/admin/wechat/config/list','ChatLineRound',3,1,NOW(),NOW()),
(40,'微信配置保存',3,39,'','POST:/admin/wechat/config/save','',1,1,NOW(),NOW()),(41,'微信配置删除',3,39,'','POST:/admin/wechat/config/delete','',2,1,NOW(),NOW()),
(42,'支付配置',2,23,'/config/channel/payment','GET:/admin/payment/config/list','Wallet',4,1,NOW(),NOW()),
(43,'支付配置保存',3,42,'','POST:/admin/payment/config/save','',1,1,NOW(),NOW()),(44,'支付配置删除',3,42,'','POST:/admin/payment/config/delete','',2,1,NOW(),NOW());
INSERT IGNORE INTO `sys_role_menu` (`role_id`,`menu_id`)
SELECT 2,`id` FROM `sys_menu` WHERE `id` IN (22,23) OR `id` BETWEEN 32 AND 44;
