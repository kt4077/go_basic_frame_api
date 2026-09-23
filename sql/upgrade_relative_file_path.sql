-- 已有数据库升级：业务表中的文件字段仅保存相对路径，完整访问地址动态补全。
-- 执行前请先备份数据库。全新安装只执行 schema.sql。

ALTER TABLE `sys_user`
  MODIFY COLUMN `avatar` VARCHAR(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
  NOT NULL DEFAULT '' COMMENT '头像相对路径';
