-- =============================================================
-- 增量升级脚本：文件上传记录表（已初始化过的库执行本文件即可）
-- 全新安装无需执行（schema.sql 已包含）。
-- =============================================================

CREATE TABLE IF NOT EXISTS `sys_upload_file` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `created_at`    DATETIME(3) NULL DEFAULT NULL,
  `updated_at`    DATETIME(3) NULL DEFAULT NULL,
  `file_name`     VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `hash_name`     VARCHAR(128) NOT NULL COMMENT '平台hash名(md5+扩展名)',
  `size`          BIGINT NOT NULL DEFAULT 0 COMMENT '文件大小(字节)',
  `mime_type`     VARCHAR(128) NOT NULL DEFAULT '' COMMENT '文件类型(MIME)',
  `ext`           VARCHAR(32) NOT NULL DEFAULT '' COMMENT '文件扩展名',
  `client`        VARCHAR(16) NOT NULL DEFAULT '' COMMENT '来源: admin/api',
  `relative_path` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '文件相对路径(云端为对象Key)',
  `absolute_path` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '上传时完整访问地址',
  `uploader_id`   BIGINT UNSIGNED NULL COMMENT '上传人ID(可空)',
  PRIMARY KEY (`id`),
  KEY `idx_sys_upload_file_hash_name` (`hash_name`),
  KEY `idx_sys_upload_file_uploader_id` (`uploader_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='上传文件记录表';
