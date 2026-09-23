# server_api 后端服务

Go + Gin + GORM(MySQL) + Redis + JWT 实现的后端基础框架。

## 目录结构

```
server_api/
├── main.go                  # 入口
├── cmd/                     # 子命令（admin / api / version）
├── config/                  # 配置加载
├── config.yaml              # 配置文件
├── sql/schema.sql           # 建表与初始数据（手动执行）
├── router/
│   ├── admin.go             # 管理端路由（独立文件）
│   └── api.go               # 用户端路由（独立文件）
├── internal/
    ├── common/              # admin/api 共用的业务能力，必须按功能使用二级目录
    │   ├── app/             #   应用依赖与连接生命周期
    │   ├── auth/            #   共用登录会话与 JWT 主动失效
    │   ├── enums/           #   共用业务枚举
    │   ├── middleware/      #   两端共用的 CORS / JWT 中间件
    │   ├── model/           #   两端共用的数据库模型
    │   └── upload/          #   共用文件上传业务
    ├── admin/               # 管理端模块（可独立迁移）
    │   ├── param/           #   请求参数结构体（按功能分文件）
    │   ├── resp/            #   响应参数结构体（按功能分文件）
    │   ├── logic/           #   业务逻辑层
    │   ├── controller/      #   控制器（绑定参数 + 调 logic + 响应）
    │   ├── middleware/      #   admin 专属中间件
    │   └── permission/      #   admin RBAC 权限业务
│   └── api/                 # 用户端模块（可独立迁移），分层同上
└── pkg/
    ├── dberror/             # 数据库错误识别
    ├── pagination/          # 分页参数规范化
    ├── password/            # 无业务归属的密码工具
    ├── response/            # Gin 统一响应工具
    ├── tree/                # 泛型树工具
    └── oss/                 # 存储平台适配器
```

## 分层约定

- controller 只做参数绑定与响应，业务逻辑全部在 logic 层实现。
- controller 调用 logic 方法时必须传入 `*gin.Context`，登录信息（claims）、IP、UA 等上下文统一在 logic 层获取。
- 请求参数结构体放 `param` 包，响应参数结构体放 `resp` 包，均不写在 controller 内。
- 每个模块的 `param`、`resp` 必须继续按功能拆分文件，禁止重新聚合成单个 `param.go`/`resp.go`。
- 两端共用的模型、枚举和业务能力放 `internal/common/<功能>`；只属于一端的能力放回对应模块。
- 数据格式化、树、密码、HTTP 响应等非业务辅助能力放 `pkg/<功能>`，不放在 `internal/common` 根目录。
- 中间件按模块归属；两端都使用的中间件放 `internal/common/middleware`。
- 管理端与用户端代码完全分模块，路由也分文件，便于独立迁移部署。

## 启动

```bash
# 1. 执行 SQL 初始化数据库
mysql -uroot -p < sql/schema.sql

# 2. 按需修改 config.yaml（MySQL/Redis/JWT）

# 3. 启动管理端 API（默认 :8001）
go run main.go service admin

# 4. 启动用户端 API（默认 :8002）
go run main.go service api

# 编译为二进制后同样支持子命令
go build -o server_api .
./server_api service admin -c config.yaml
./server_api service api -c config.yaml
```

服务启动时会检查 MySQL、Redis 连接及必需数据表；收到 SIGINT/SIGTERM 后会在
`server.shutdown_timeout_seconds` 时间内优雅停止 HTTP 服务，再关闭连接池。
HTTP 读写、请求头及空闲连接超时均可在 `config.yaml` 的 `server` 节点配置。

已有数据库升级时按需手动执行 `sql/upgrade_*.sql`。并发一致性约束见
`sql/upgrade_concurrency_constraints.sql`，它会为角色编码增加唯一约束，并保证
同一时间最多只有一个未删除的默认存储渠道。

短信、微信、支付渠道配置的已有库升级脚本为 `sql/upgrade_channel_configs.sql`。
短信服务商、模板类型、发送状态、微信应用类型和支付渠道等数值枚举统一从 1 开始。

管理员头像字段的已有库升级脚本为 `sql/upgrade_user_avatar.sql`；全新数据库无需执行，
`schema.sql` 已包含该字段。

头像等业务字段仅保存相对路径，完整访问地址按当前默认存储配置动态生成；
`sys_upload_file` 作为上传档案表，同时保存相对路径和上传时的完整地址快照。

## 登录鉴权设计（JWT 主动过期）

1. 登录成功：生成 `login_id`(uuid) 写入 `sys_user_login` 流水表，并签发携带 `login_id` 的 JWT。
2. 同时将 `login:<login_id>` 写入 Redis（TTL 与 JWT 有效期一致）。
3. 每次请求：解析 JWT 后检查 Redis 中 `login:<login_id>` 是否存在，不存在即拒绝 —— 实现 JWT 主动过期。
4. 主动退出 / 管理端禁用账号 / 踢人下线 / 修改密码：删除对应 Redis 缓存键，该账号所有会话立即失效。

## 接口权限

- 菜单/按钮（`sys_menu.type=3` 的按钮）通过 `api_path` 绑定后端接口，格式 `METHOD:/path`，如 `POST:/admin/user/add`。
- 用户接口权限 = 其角色（含后代角色的继承权限）绑定的全部菜单/按钮 `api_path` 集合，缓存于 Redis `perm:<user_id>`。
- 中间件按 `请求方法:路由pattern` 与权限集合匹配；超级管理员直接放行。
- 菜单/角色/用户权限变更后自动清除权限缓存。
- 权限不匹配返回 403；Redis 或权限数据源异常返回 503 并写入服务日志，避免把系统故障误报为无权限。

## 文件安全

- 单文件上限为 50MB，存储层通过临时文件流式计算 hash，不会把整个文件载入内存。
- 远程文件抓取仅允许 HTTP(S)，限制请求和重定向时间，并拒绝内网、回环、链路本地及保留地址，防止 SSRF。

## 初始账号

| 账号 | 密码 | 说明 |
| ---- | ---- | ---- |
| admin | 123456 | 超级管理员，拥有全部权限 |
| zhangsan | 123456 | 运营专员，仅系统总览 |
