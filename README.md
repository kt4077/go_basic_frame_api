# server_api

基于 Go 的后台管理 / 用户端**双端 API 基础框架**：Gin + GORM(MySQL) + Redis + JWT，内置 RBAC 权限、菜单/部门/角色树、操作日志、统一文件上传（本地 / 阿里云 OSS / 腾讯云 COS / 七牛云 Kodo / MinIO）以及短信、微信、支付渠道配置能力。

管理端与用户端代码、路由、启动命令完全分离，可独立部署或拆分迁移。

---

## 一、技术栈

| 类别 | 选型 | 说明 |
| --- | --- | --- |
| 语言 | Go 1.26（`module server_api`） | 无 CGO 依赖，可交叉编译 |
| Web 框架 | gin-gonic/gin v1.10 | 路由分组 + 中间件链 |
| ORM | gorm.io/gorm v1.31 + mysql 驱动 | 不使用 AutoMigrate，表结构由 SQL 脚本维护 |
| 缓存 / 会话 | redis/go-redis v9 | 登录态、权限缓存 |
| 鉴权 | golang-jwt/jwt v5（HS256） | JWT + Redis 实现主动过期 |
| CLI | spf13/cobra | `service admin` / `service api` / `version` 子命令 |
| 配置 | gopkg.in/yaml.v3 | 单文件 YAML，直接反序列化 |
| 密码 | golang.org/x/crypto/bcrypt | 默认 cost |
| 对象存储 | 阿里云 OSS / 腾讯云 COS / 七牛云 Kodo / MinIO SDK | 统一 `Uploader` 接口 |
| 其他 | google/uuid | 生成 `login_id` |

日志使用标准库 `log` + Gin 自带 `Logger/Recovery`，GORM 开启慢 SQL（1s）告警；管理端业务操作写入 `sys_operation_log`。

---

## 二、目录结构

```
server_api/
├── main.go                  # 入口：执行 cmd.RootCmd()
├── cmd/                     # Cobra 子命令（service admin / service api / version）
├── config/
│   └── config.go            # 配置结构体、加载与默认值
├── config.yaml              # 运行时配置（含敏感信息，生产环境请勿入库）
├── config.example.yaml      # 配置模板（占位值，可安全提交）
├── router/
│   ├── admin.go             # 管理端路由（监听 server.admin_addr）
│   └── api.go               # 用户端路由（监听 server.api_addr）
├── internal/
│   ├── common/              # 两端共用能力
│   │   ├── app/             #   配置 + GORM + Redis 的初始化、健康检查、连接关闭
│   │   ├── auth/            #   登录流水、JWT 签发/解析、踢人下线
│   │   ├── enums/           #   业务枚举（状态、菜单类型、渠道、权限白名单等）
│   │   ├── middleware/      #   CORS、JWT 登录鉴权
│   │   ├── model/           #   GORM 数据模型（sys_* 表）
│   │   └── upload/          #   上传业务、默认存储加载、本地文件访问
│   ├── admin/               # 管理端模块
│   │   ├── controller/      #   参数绑定 + 调 logic + 响应
│   │   ├── logic/           #   业务逻辑（必须接收 *gin.Context）
│   │   ├── middleware/      #   Permission（接口级权限）、OperationLog（操作日志）
│   │   ├── param/           #   请求参数结构体（按功能分文件）
│   │   ├── resp/            #   响应结构体（按功能分文件）
│   │   └── permission/      #   RBAC 权限计算与缓存
│   └── api/                 # 用户端模块（分层同上）
├── pkg/
│   ├── dberror/             # 数据库错误识别（如重复键 1062）
│   ├── oss/                 # 多平台对象存储适配器（单文件上限 50MB）
│   ├── pagination/          # 分页参数规范化（默认第 1 页、20 条，最大 100）
│   ├── password/            # bcrypt 哈希与校验
│   ├── response/            # 统一 JSON 响应封装
│   └── tree/                # 泛型树工具（菜单/部门/角色树）
├── sql/                     # 数据库脚本目录（建表 / 升级脚本）
└── uploads/                 # 本地存储根目录（运行时生成，已忽略入库）
```

分层约定：

- controller 只负责绑定参数与响应，业务逻辑一律下沉到 logic。
- controller 调用 logic 时必须传入 `*gin.Context`，claims、IP、UA 等上下文统一在 logic 内获取。
- 请求/响应结构体分别放 `param`、`resp`，且必须按功能拆分文件，禁止合并为单个 `param.go`。
- 两端共用能力放 `internal/common/<功能>`；无业务归属的工具放 `pkg/<功能>`。

---

## 三、快速开始

### 1. 环境要求

- Go 1.26+
- MySQL 5.7+ / 8.0+（字符集 `utf8mb4`）
- Redis 5+

### 2. 初始化数据库

```bash
# 执行建表脚本（脚本由部署包提供或放置于 sql/ 目录）
mysql -uroot -p < sql/schema.sql

# 已有库升级：按需执行对应的 upgrade_*.sql
```

服务启动时会校验 16 张必需表是否齐全，缺表会拒绝启动并提示：

```
缺少数据表 [...]，请先执行 sql/schema.sql 初始化数据库
```

同时要求 `sys_user` 至少存在一条记录，否则同样拒绝启动。

必需表清单：

```
sys_dept  sys_user  sys_role  sys_menu  sys_user_login  sys_user_role  sys_role_menu
sys_operation_log  sys_storage_config  sys_upload_file
sys_sms_config  sys_sms_signature  sys_sms_template  sys_sms_send_log
sys_wechat_config  sys_payment_config
```

> 仓库当前 `sql/` 目录未内置脚本，请从部署包获取或使用自己维护的建表脚本；表结构可参照 `internal/common/model` 中的模型定义。

### 3. 修改配置

```bash
cp config.example.yaml config.yaml
vim config.yaml   # 填写 MySQL / Redis / JWT
```

### 4. 启动服务

```bash
# 管理端 API（默认 :8001）
go run main.go service admin

# 用户端 API（默认 :8002）
go run main.go service api
```

### 5. 编译与运行

```bash
go build -o server_api .

./server_api service admin -c config.yaml
./server_api service api   -c config.yaml
```

---

## 四、服务管理

### 子命令

| 命令 | 说明 |
| --- | --- |
| `server_api service admin` | 启动管理端 API，监听 `server.admin_addr` |
| `server_api service api` | 启动用户端 API，监听 `server.api_addr` |
| `server_api version` | 输出版本号 |
| `server_api help` | 查看命令帮助 |

两个 service 子命令均支持 `-c/--config`，默认读取当前目录下的 `config.yaml`：

```bash
./server_api service admin -c /etc/server_api/config.yaml
```

### 启动自检

启动按顺序执行：

1. 读取并校验 YAML 配置（`admin_addr`、`api_addr` 必填，连接池参数需合法）；
2. 打开 MySQL 并在 5 秒内 `ping`，配置连接池（最大生命周期 30 分钟、最大空闲 5 分钟）；
3. 创建 Redis 客户端并在 5 秒内 `PING`（失败时同时关闭已打开的 MySQL 连接）；
4. 校验必需数据表；
5. 校验 `sys_user` 非空；
6. 按模式加载管理端或用户端路由并监听端口。

任一步失败即中止启动并返回明确错误。

### 优雅退出

- 监听 `SIGINT` / `SIGTERM`；
- 收到信号后在 `server.shutdown_timeout_seconds`（默认 15 秒）内调用 `http.Server.Shutdown`，停止接收新连接并等待在途请求完成；
- 随后先关闭 Redis，再关闭 MySQL 连接池；
- 超时或失败会输出「服务优雅停机失败」。

### 进程守护示例

```bash
# nohup 后台运行
nohup ./server_api service admin -c config.yaml >> admin.log 2>&1 &

# systemd 示例
[Unit]
Description=server_api admin
After=network.target mysql.service redis.service

[Service]
Type=simple
WorkingDirectory=/opt/server_api
ExecStart=/opt/server_api/server_api service admin -c /opt/server_api/config.yaml
Restart=always
RestartSec=3
KillSignal=SIGTERM

[Install]
WantedBy=multi-user.target
```

---

## 五、配置说明

完整模板见 `config.example.yaml`。

| 配置项 | 说明 | 默认值 / 校验 |
| --- | --- | --- |
| `server.admin_addr` | 管理端监听地址 | 必填，如 `:8001` |
| `server.api_addr` | 用户端监听地址 | 必填，如 `:8002` |
| `server.read_header_timeout_seconds` | 读取请求头超时 | 缺省 5 |
| `server.read_timeout_seconds` | 请求读取超时 | 缺省 60 |
| `server.write_timeout_seconds` | 响应写入超时 | 缺省 60 |
| `server.idle_timeout_seconds` | 空闲连接超时 | 缺省 120 |
| `server.shutdown_timeout_seconds` | 优雅停机期限 | 缺省 15 |
| `mysql.host/port/user/password/database` | MySQL 连接信息 | 必填 |
| `mysql.max_open_conns` | 最大连接数 | 必须 > 0 |
| `mysql.max_idle_conns` | 最大空闲连接数 | ≥ 0 且 ≤ 最大连接数 |
| `redis.addr/password/db` | Redis 连接信息 | 密码为空表示无密码 |
| `jwt.secret` | HS256 签名密钥 | 生产必须更换，可用 `openssl rand -hex 32` |
| `jwt.expire_hours` | Token 与登录缓存有效期（小时） | 缺省配置 24 |
| `jwt.issuer` | 签发方标识 | 如 `go_backend_frame` |
| `log.level` | 日志级别 `debug/info/warn/error` | `info` |

DSN 固定追加 `charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local`。

---

## 六、接口一览

统一响应格式由 `pkg/response` 封装，业务码：`0` 成功、`400` 参数错误、`401` 未登录、`403` 无权限、`500` 服务异常、`503` 依赖不可用（如 Redis 异常）。

### 管理端（`:8001`，共 55 条）

公共中间件：`gin.Logger → gin.Recovery → CORS`；受保护路由追加 `Auth → Permission → OperationLog`（部分仅 `Auth`）。

| 模块 | 方法 | 路径 | 鉴权 |
| --- | --- | --- | --- |
| 文件访问 | GET | `/files/*filepath` | 公开 |
| 登录 | POST | `/admin/login` | 公开 |
| 登录会话 | POST | `/admin/logout` | 登录 |
| 登录会话 | GET | `/admin/me` | 登录 |
| 登录会话 | GET | `/admin/routers` | 登录 |
| 登录会话 | GET | `/admin/permissions` | 登录 |
| 登录会话 | POST | `/admin/profile/update` | 登录 |
| 登录会话 | POST | `/admin/change_password` | 登录 |
| 文件上传 | POST | `/admin/upload/file` | 登录（权限白名单） |
| 系统总览 | GET | `/admin/dashboard/overview` | 登录 + 权限 |
| 菜单管理 | GET | `/admin/menu/list` | 登录 + 权限 |
| 菜单管理 | GET | `/admin/menu/tree` | 登录 + 权限 |
| 菜单管理 | POST | `/admin/menu/add` | 登录 + 权限 |
| 菜单管理 | POST | `/admin/menu/update` | 登录 + 权限 |
| 菜单管理 | POST | `/admin/menu/delete` | 登录 + 权限 |
| 角色管理 | GET | `/admin/role/list` | 登录 + 权限 |
| 角色管理 | GET | `/admin/role/tree` | 登录 + 权限 |
| 角色管理 | POST | `/admin/role/add` | 登录 + 权限 |
| 角色管理 | POST | `/admin/role/update` | 登录 + 权限 |
| 角色管理 | POST | `/admin/role/delete` | 登录 + 权限 |
| 角色管理 | GET | `/admin/role/menus?id=` | 登录 + 权限 |
| 角色管理 | POST | `/admin/role/assign_menus` | 登录 + 权限 |
| 角色管理 | GET | `/admin/role/users?id=` | 登录 + 权限 |
| 部门管理 | GET | `/admin/dept/tree` | 登录 + 权限 |
| 部门管理 | POST | `/admin/dept/add` | 登录 + 权限 |
| 部门管理 | POST | `/admin/dept/update` | 登录 + 权限 |
| 部门管理 | POST | `/admin/dept/delete` | 登录 + 权限 |
| 用户管理 | GET | `/admin/user/list` | 登录 + 权限 |
| 用户管理 | POST | `/admin/user/add` | 登录 + 权限 |
| 用户管理 | POST | `/admin/user/update` | 登录 + 权限 |
| 用户管理 | POST | `/admin/user/delete` | 登录 + 权限 |
| 用户管理 | POST | `/admin/user/reset_password` | 登录 + 权限 |
| 用户管理 | POST | `/admin/user/kick` | 登录 + 权限 |
| 操作日志 | GET | `/admin/log/operation/list` | 登录 + 权限 |
| 存储渠道 | GET | `/admin/storage/list` | 登录 + 权限 |
| 存储渠道 | POST | `/admin/storage/add` | 登录 + 权限 |
| 存储渠道 | POST | `/admin/storage/update` | 登录 + 权限 |
| 存储渠道 | POST | `/admin/storage/set_default` | 登录 + 权限 |
| 存储渠道 | POST | `/admin/storage/delete` | 登录 + 权限 |
| 短信配置 | GET | `/admin/sms/config/list` | 登录 + 权限 |
| 短信配置 | POST | `/admin/sms/config/save` | 登录 + 权限 |
| 短信配置 | POST | `/admin/sms/config/delete` | 登录 + 权限 |
| 短信签名 | GET | `/admin/sms/signature/list` | 登录 + 权限 |
| 短信签名 | POST | `/admin/sms/signature/save` | 登录 + 权限 |
| 短信签名 | POST | `/admin/sms/signature/delete` | 登录 + 权限 |
| 短信模板 | GET | `/admin/sms/template/list` | 登录 + 权限 |
| 短信模板 | POST | `/admin/sms/template/save` | 登录 + 权限 |
| 短信模板 | POST | `/admin/sms/template/delete` | 登录 + 权限 |
| 短信记录 | GET | `/admin/sms/log/list` | 登录 + 权限 |
| 微信配置 | GET | `/admin/wechat/config/list` | 登录 + 权限 |
| 微信配置 | POST | `/admin/wechat/config/save` | 登录 + 权限 |
| 微信配置 | POST | `/admin/wechat/config/delete` | 登录 + 权限 |
| 支付配置 | GET | `/admin/payment/config/list` | 登录 + 权限 |
| 支付配置 | POST | `/admin/payment/config/save` | 登录 + 权限 |
| 支付配置 | POST | `/admin/payment/config/delete` | 登录 + 权限 |

### 用户端（`:8002`，共 6 条）

| 模块 | 方法 | 路径 | 鉴权 |
| --- | --- | --- | --- |
| 文件访问 | GET | `/files/*filepath` | 公开 |
| 登录 | POST | `/api/login` | 公开 |
| 登录会话 | POST | `/api/logout` | 登录 |
| 用户资料 | GET | `/api/profile` | 登录 |
| 用户资料 | POST | `/api/change_password` | 登录 |
| 文件上传 | POST | `/api/upload/file` | 登录 |

用户端不使用接口级 RBAC 与操作日志中间件，仅需通过 JWT 登录鉴权。

### 中间件

| 中间件 | 作用域 | 说明 |
| --- | --- | --- |
| `gin.Logger` / `gin.Recovery` | 两端 | 访问日志与 panic 恢复 |
| `CORS` | 两端 | 回显 Origin，允许 `GET,POST,PUT,DELETE,PATCH,OPTIONS`，OPTIONS 返回 204 |
| `Auth` | 两端受保护路由 | 解析 `Authorization: Bearer <token>`，校验签名与 Redis 登录态 |
| `Permission` | 管理端 | 接口级 RBAC 校验，超级管理员放行，无权限 403、依赖异常 503 |
| `OperationLog` | 管理端 | 记录方法、路由、参数、响应、IP、UA、耗时、操作人；对 `password/token/secret/*_key` 等字段脱敏 |

---

## 七、鉴权与权限

### JWT 主动过期

1. 登录成功生成 `login_id`(UUID) 写入 `sys_user_login`，签发携带 `login_id` 的 JWT（HS256，Claims 含 `user_id/username/login_id/is_super/client/iss/exp/iat`）；
2. 同时写入 Redis `login:<login_id>`，TTL 与 JWT 有效期一致；
3. 每次请求解析 JWT 后检查该 Key，不存在即判定登录失效（即使 JWT 未到期）；
4. 退出、踢人、禁用账号、改密、重置密码、删除用户时删除对应 Key，会话立即失效。

### 权限模型

```
用户 ── sys_user_role ── 角色 ── sys_role_menu ── 菜单/按钮(sys_menu.api_path)
```

- 菜单类型：`1` 目录、`2` 菜单、`3` 按钮；按钮通过 `api_path` 绑定接口，格式 `METHOD:/path`（如 `POST:/admin/user/add`），支持逗号分隔多个。
- 角色为树形结构，上级角色自动继承全部后代角色的权限。
- 权限集合缓存于 Redis `perm:<user_id>`（TTL 24 小时，空集合用 `__none__` 占位）；菜单、角色、用户权限变更时自动清除。
- 超级管理员（`is_super=1`）直接放行；超级管理员不可被踢下线或删除。

### 初始账号

| 账号 | 密码 | 说明 |
| --- | --- | --- |
| admin | 123456 | 超级管理员，拥有全部权限 |
| zhangsan | 123456 | 运营专员，仅系统总览 |

---

## 八、文件上传与存储

- 单文件上限 **50MB**；上传时流式写入临时文件并计算 MD5，不会整文件载入内存。
- 对象 Key 规则：`uploads/YYYY/MM/DD/<md5>.<ext>`；本地存储根目录 `./uploads`。
- 存储渠道由数据库 `sys_storage_config` 动态配置（唯一默认渠道），支持 `local`、`aliyun`、`tencent`、`qiniu`、`minio`（S3 兼容）。
- 业务字段仅保存相对路径，完整访问地址按当前默认存储配置动态生成；`sys_upload_file` 作为上传档案表，同时保存相对路径与上传时的完整地址快照。
- 远程文件抓取仅允许 HTTP(S)，限制超时并拒绝内网、回环、链路本地与保留地址，防止 SSRF。

---

## 九、运维与安全建议

- `config.yaml` 含数据库密码与 JWT 密钥，请勿提交；使用 `config.example.yaml` 作为模板，并确保 `jwt.secret` 在生产环境替换为随机值。
- 前端需配置代理：管理端 `/admin` → `http://127.0.0.1:8001`，用户端 `/api` → `http://127.0.0.1:8002`。
- Redis 不可用时，管理端接口返回 503 而非 403，避免把依赖故障误报为无权限。
- 数据库不使用 AutoMigrate，表结构变更请通过 `sql/` 下的升级脚本执行。
- 生产环境建议在反向代理后运行，并仅暴露必要的 `/files/*filepath` 静态访问路径。
