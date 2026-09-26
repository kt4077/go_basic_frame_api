# 短信渠道接入说明

核心框架统一通过 `pkg/sms` 发送短信，管理端继续使用现有的“开发信息、签名模板、短信模板、发送记录”页面。本次扩展不新增数据表，也不改变现有表结构。

## 支持渠道

| provider | 渠道 | `access_key_id` | `access_key_secret` | 模板编码 | 默认 Endpoint |
| --- | --- | --- | --- | --- | --- |
| 1 | 阿里云 | AccessKey ID | AccessKey Secret | 必填 | `https://dysmsapi.aliyuncs.com` |
| 2 | 腾讯云 | 短信应用 SDK AppID | Secret Key | 必填 | `https://sms.tencentcloudapi.com` |
| 3 | 短信宝 | 登录用户名 | API Key 或 32 位 MD5 密码 | 可选，作为产品 ID `g` | `https://api.smsbao.com/sms` |
| 4 | SMS.cn | 登录账号 `uid` | `md5(登录密码 + uid)` 得到的 32 位接口密码 | 必填 | `https://api.sms.cn/sms/` |
| 5 | 云片 | API Key | 不使用 | 可留空 | `https://sms.yunpian.com/v2/sms/single_send.json` |

管理端不会回传已保存的密钥。修改配置时密钥留空表示继续使用原值；云片不需要填写密钥字段。自定义 Endpoint 仅允许 `http` 或 `https` 地址，生产环境应使用 HTTPS 并限制配置权限。

## 签名与模板

- 阿里云、腾讯云和 SMS.cn 使用平台审核通过的模板编码。
- 短信宝和云片使用系统模板填充后的完整正文；框架会在正文前拼接 `【签名】`，如果正文已经以该签名开头则不会重复拼接。
- 短信宝的模板编码可填写平台产品 ID；没有使用产品 ID 时可以留空。
- 验证码模板支持 `${code}`、`{code}`、`{1}` 和 `#{code}` 占位符。
- 当前发送逻辑按 ID 升序选择第一条启用的开发配置，以及该配置下第一条启用的签名和验证码模板。需要切换渠道时，应确保目标配置启用，并停用优先级更高的配置。

## 请求与发送记录

三家新增渠道的请求约定如下：

- 短信宝：GET 请求，成功响应码为 `0`。
- SMS.cn：表单 POST 请求，成功响应字段为 `stat=100`；验证码参数按 `{"code":"123456"}` 发送。
- 云片：表单 POST 请求，成功响应字段为 `code=0`，返回的 `sid` 会写入短信发送记录的平台消息 ID。

所有渠道都复用现有的请求超时、验证码限流和发送日志机制。渠道原始密钥不会写入日志；对外接口只返回统一错误，渠道错误详情仅保存在管理端发送记录中。

## 配置与验证步骤

1. 在“系统配置 / 短信配置 / 开发信息”新增渠道凭证，Endpoint 通常留空。
2. 为该开发配置新增已审核的签名。
3. 新增类型为“验证码”的模板，并填写模板内容；按上表决定是否填写模板编码。
4. 确保待验证渠道是 ID 最小的启用配置，其他启用配置暂时停用。
5. 调用用户端 `POST /api/sms/code`，在发送记录中检查状态、错误信息和平台消息 ID。

## 官方文档

- [短信宝 HTTP API](https://www.smsbao.com/openapi/213.html)
- [SMS.cn 短信 API](https://www.sms.cn/sms_api.html)
- [云片国内短信 API](https://www.yunpian.com/official/document/sms/zh_CN/domestic_list)
- [云片单条发送接口](https://www.yunpian.com/official/document/sms/zh_cn/domestic_single_send)
