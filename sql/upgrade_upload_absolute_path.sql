-- 若曾执行过删除 sys_upload_file.absolute_path 的过渡脚本，使用本脚本恢复上传地址快照字段。
-- 全新安装或表中已存在 absolute_path 时无需执行。

ALTER TABLE `sys_upload_file`
  ADD COLUMN `absolute_path` VARCHAR(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci
  NOT NULL DEFAULT '' COMMENT '上传时完整访问地址' AFTER `relative_path`;
