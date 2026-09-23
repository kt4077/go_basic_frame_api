-- 已有数据库升级：管理员头像字段。全新安装只执行 schema.sql。
ALTER TABLE `sys_user`
  ADD COLUMN `avatar` VARCHAR(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '头像相对路径' AFTER `nickname`;
