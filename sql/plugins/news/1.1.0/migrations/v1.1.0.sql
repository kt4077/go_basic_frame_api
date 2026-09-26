ALTER TABLE `plg_news_article`
  ADD COLUMN `enable_status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '启用状态，1启用，2禁用' AFTER `status`,
  ADD COLUMN `virtual_view_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '虚拟浏览次数' AFTER `view_count`;
