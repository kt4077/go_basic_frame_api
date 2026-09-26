CREATE TABLE IF NOT EXISTS `plg_news_category` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '分类名称',
  `slug` varchar(64) COLLATE utf8mb4_general_ci NOT NULL COMMENT '分类标识',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态，1启用，2停用',
  `remark` varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '分类备注',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plg_news_category_slug` (`slug`),
  KEY `idx_plg_news_category_sort` (`sort`),
  KEY `idx_plg_news_category_status` (`status`),
  KEY `idx_plg_news_category_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='新闻分类表';

CREATE TABLE IF NOT EXISTS `plg_news_article` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `category_id` bigint unsigned NOT NULL COMMENT '分类ID',
  `title` varchar(200) COLLATE utf8mb4_general_ci NOT NULL COMMENT '文章标题',
  `slug` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '文章标识',
  `summary` varchar(500) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '文章摘要',
  `cover` varchar(1024) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '封面相对路径',
  `content` longtext COLLATE utf8mb4_general_ci NOT NULL COMMENT '文章正文',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态，1草稿，2已发布，3已下线',
  `sort` int NOT NULL DEFAULT '0' COMMENT '排序值',
  `view_count` bigint unsigned NOT NULL DEFAULT '0' COMMENT '浏览次数',
  `published_at` datetime DEFAULT NULL COMMENT '发布时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at` datetime DEFAULT NULL COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_plg_news_article_slug` (`slug`),
  KEY `idx_plg_news_article_category_status` (`category_id`,`status`),
  KEY `idx_plg_news_article_sort` (`sort`),
  KEY `idx_plg_news_article_published_at` (`published_at`),
  KEY `idx_plg_news_article_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='新闻文章表';
