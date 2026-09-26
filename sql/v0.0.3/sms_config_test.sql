-- v0.0.3 短信渠道配置测试接口权限升级脚本。
-- 复用“短信开发配置保存”按钮权限，避免为已有角色新增独立授权记录。
-- 兼容 MySQL 5.7/8.0；可重复执行，不指定任何业务数据主键。

UPDATE `sys_menu`
SET `api_path` = CONCAT(`api_path`, ',POST:/admin/sms/config/test')
WHERE `type` = 3
  AND FIND_IN_SET('POST:/admin/sms/config/save', REPLACE(`api_path`, ' ', '')) > 0
  AND FIND_IN_SET('POST:/admin/sms/config/test', REPLACE(`api_path`, ' ', '')) = 0;
