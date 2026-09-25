# 代码规范说明

适用范围：`go_backend_frame` 仓库，包含后端服务 `server_api`（Go）与管理前端 `admin_client`（Vue 3 + TypeScript）。

目标：让新增代码与现有工程保持一致的目录结构、分层方式、命名风格与安全边界，降低协作与维护成本。

---

## 一、仓库总览

```
go_backend_frame/
├── server_api/       # 后端：Go + Gin + GORM(MySQL) + Redis + JWT（本目录）
│   ├── cmd/          # Cobra 子命令：service admin / service api / version
│   ├── config/       # 配置结构体与加载（含默认值）
│   ├── router/       # admin.go（管理端）、api.go（用户端），两端路由分文件
│   ├── internal/     # 业务代码（对外不可引用）
│   │   ├── common/   # 两端共用：app / auth / enums / middleware / model / upload
│   │   ├── admin/    # 管理端：controller / logic / param / resp / middleware / permission
│   │   └── api/      # 用户端：controller / logic / param / resp
│   ├── pkg/          # 无业务归属的工具：response / pagination / password / tree / dberror / oss / mask
│   ├── sql/          # 数据库脚本（建表与升级）
│   └── docs/         # 接口文档与本规范
└── admin_client/     # 前端：Vue 3 + Vite + TS + Element Plus
    └── src/
        ├── api/        # 按业务模块封装后端请求
        ├── types/      # 请求与响应类型
        ├── enums/      # 前端业务枚举（值与后端一致）
        ├── utils/      # 认证、日期、图表等工具
        ├── store/      # Pinia 状态
        ├── router/     # 静态路由与动态路由注册
        ├── directives/ # v-perm 等指令
        ├── components/ # 通用组件
        ├── styles/     # 全局主题与 UI 规范
        └── views/      # 页面模块
```

---

## 二、后端 Go 规范

### 2.1 分层职责（强制）

| 层 | 目录 | 职责 | 禁止 |
| --- | --- | --- | --- |
| 路由 | `router/` | 注册路由与中间件链，按端分文件 | 写业务逻辑 |
| 控制器 | `internal/<端>/controller/` | 绑定参数 → 调 logic → 响应 | 写 SQL、写业务规则 |
| 业务 | `internal/<端>/logic/` | 业务规则、事务、缓存维护 | 直接写 HTTP 响应 |
| 参数 | `internal/<端>/param/` | 请求结构体（按功能拆分文件） | 与 resp 混写 |
| 响应 | `internal/<端>/resp/` | 响应结构体（按功能拆分文件） | 在 controller 内定义匿名结构 |
| 模型 | `internal/common/model/` | GORM 模型（`sys_*` 表） | 在 resp 中定义表结构 |
| 中间件 | `common/middleware`、`admin/middleware` | 横切逻辑（CORS / 鉴权 / 权限 / 日志） | 业务分支 |
| 工具 | `pkg/` | 无业务归属的通用能力 | 依赖 internal |

核心约定：

- controller 方法只做三件事：`ShouldBind` → 调 logic → `response.OK/Fail`。
- controller 调用 logic **必须传入 `*gin.Context`**；登录信息（claims）、IP、User-Agent 统一在 logic 层通过 `auth.CtxClaims(c)`、`c.ClientIP()`、`c.GetHeader("User-Agent")` 获取。
- `param`、`resp` 必须按功能拆分为多个文件（如 `user.go`、`menu.go`），**禁止**合并成单个 `param.go` / `resp.go`。
- 两端共用能力放 `internal/common/<功能>`；只属于一端的能力放回该端目录；无业务归属的工具放 `pkg/<功能>`。
- 管理端与用户端代码完全分模块，便于独立迁移部署。

### 2.2 命名规范

- 文件名：小写 + 下划线（`operation_log.go`、`sys_user_login.go`）。
- 包名：简短单数名词，与目录一致（`param`、`resp`、`logic`）。
- 结构体：大驼峰；请求以 `Req` 结尾，响应以 `Res` / `Item` 结尾（`UserSaveReq`、`UserItem`、`OverviewRes`）。
- 控制器以 `Controller` 结尾（`UserController`），业务以 `Logic` 结尾（`UserLogic`），字段名为 `Logic`。
- 方法名与路由动作对应：`List / Create / Update / Delete / Tree / Save / SetDefault / Kick / ResetPassword`。
- 错误文本使用中文，面向使用者可读（`errors.New("该菜单下存在子级，不能删除")`）。

### 2.3 结构体 Tag 规范

- JSON 字段统一 `snake_case`：`json:"parent_id"`。
- 查询参数接口（GET）字段同时声明 `form:"..."`；请求体接口（POST）只需 `json:"..."`。
- 校验使用 `binding`：`required`、`min=6`、`oneof=1 2`、`omitempty,max=1024`、`url`、`email`。
- 每个字段写 `comment:"..."` 中文注释，作为接口文档与前端类型的事实来源。

```go
type UserSaveReq struct {
    ID       uint   `json:"id" comment:"主键ID"`
    Username string `json:"username" comment:"登录账号"`
    Password string `json:"password" comment:"登录密码"`
    RoleIDs  []uint `json:"role_ids" comment:"角色ID列表"`
}
```

### 2.4 响应与错误

- 统一使用 `pkg/response`：`response.OK(c, data)` 与 `response.Fail(c, code, msg)`，HTTP 状态码恒为 200。
- 业务码：`0` 成功、`400` 参数错误、`401` 未登录/登录失效、`403` 无权限、`500` 业务失败、`503` 依赖不可用。
- 参数绑定失败统一返回 `CodeErrParams`，不要透传 Gin 原始错误。
- 数据库错误使用 `pkg/dberror` 识别（如 `IsDuplicateKey` → 重复提示），不要直接把驱动错误返回前端。
- 无返回数据的操作接口（删除、设置默认、踢下线）使用 `response.OK(c, nil)`。

### 2.5 数据访问

- 不使用 `AutoMigrate`，表结构变更必须提供 `sql/` 下对应脚本（建表脚本 + `upgrade_*.sql`）。
- 查询统一走 GORM，禁止拼接 SQL 字符串；动态条件使用 `db = db.Where(...)` 链式构造。
- 批量写入、跨表一致性变更使用事务；保存类接口优先使用 `clause.OnConflict` 做 upsert，避免并发产生重复数据。
- 并发一致性依赖数据库唯一约束（如角色编码唯一、唯一默认渠道），不能只靠应用层判断。
- 软删除模型使用 `gorm.DeletedAt`，查询默认排除已删除记录。
- 金额类字段数据库使用 `decimal`（如 `sys_member.balance` 为 `decimal(12,2)`），Go 模型与 Resp 用字符串承载，禁止 float/double，避免精度丢失。

### 2.6 鉴权与权限

- 受保护路由必须挂在带 `commonmiddleware.Auth` 的路由组上。
- 管理端业务接口挂 `Auth → Permission → OperationLog`；用户端仅 `Auth`。
- 按钮权限通过 `sys_menu.api_path` 绑定，格式 `METHOD:/path`（如 `POST:/admin/user/add`），多个用逗号分隔。
- 通用接口（如文件上传）如需免接口鉴权，必须在 `enums.SkipPermissionApis` 白名单中显式登记。
- 菜单、角色、用户权限变更后，必须调用 `permission.ClearAllPermissionCache` 或清对应 `perm:<user_id>` 缓存。
- 禁止在 logic 之外绕过 Permission 中间件做权限判断；前端显隐只是体验，安全边界在后端。

### 2.7 配置与敏感信息

- 运行时配置放 `config.yaml`，模板为 `config.example.yaml`；`config.yaml` 已在 `.gitignore` 中，禁止提交。
- 新增配置项必须同步更新 `config/config.go` 结构体、`config.example.yaml` 与 README 配置表。
- 密钥类字段（私钥、公钥、APIv3 密钥、Secret）写入时允许为空表示不修改，**响应结构一律不返回**。
- 禁止把密钥、证书、真实地址写入代码或提交到仓库；证书类文件（`.pem`、`.key`）已被忽略。

### 2.8 日志与审计

- 管理端操作日志由 `OperationLog` 中间件记录方法、路由、参数、响应、IP、UA、耗时与操作人。
- 新增敏感字段（密码、token、secret、各类 key）需同步加入脱敏关键字列表。
- 查询日志类接口（如操作日志列表）应加入跳过名单，避免自身产生日志风暴。
- 服务日志使用标准库 `log` + Gin 日志，慢 SQL 阈值 1s。

### 2.9 文件上传与存储

- 单文件上限 50MB，走存储层流式处理，禁止把整个文件读入内存。
- 业务字段只保存 `relative_path`，展示时通过当前默认存储配置动态拼地址。
- `sys_upload_file` 同时保存相对路径与上传时的完整地址快照。
- 远程抓取仅允许 HTTP(S)，复用 `safeRemoteHTTPClient`，禁止自行发起未校验的外部请求（防 SSRF）。

### 2.10 枚举与常量

- 业务枚举集中在 `internal/common/enums/`，按功能分文件（菜单、渠道、登录、平台等）。
- 数值枚举**统一从 1 开始**，`0` 仅用于“未设置/不过滤”语义。
- 前端 `src/enums` 必须与后端取值保持一致。

### 2.11 注释与文档

- 每个 `package` 写一段包注释说明职责与约定（如 `param`、`resp`、`controller`、`logic`）。
- 导出方法写单行注释，说明做什么、以及与权限/缓存相关的副作用。
- 新增或修改接口后，同步更新 `docs/admin_openapi.yaml` 与 Apifox 在线文档。

### 2.12 测试与提交前检查

- 单元测试与被测文件同目录，命名 `*_test.go`，优先表驱动。
- 提交前执行：

```bash
gofmt -l .            # 不允许有输出
go vet ./...
go build ./...
go test ./...         # 有测试时
```

---

## 三、前端 Vue / TypeScript 规范

### 3.1 目录与模块划分

- 每个业务模块在 `src/api`、`src/types` 中各有独立文件，文件名与后端模块一致（`user.ts`、`sms.ts`、`storage.ts`）。
- 页面放 `src/views/<模块>/<页面>/index.vue`，与后端菜单 `path` 对应（`/system/user` → `views/system/user/index.vue`）。
- 可复用 UI 抽到 `src/components/`；横切行为用 `src/directives/`。
- 全局状态只放 `src/store/`，能由接口直接获取的数据不要进 store。

### 3.2 接口层

- 统一通过 `src/api/http.ts` 的 `request<T>()`，禁止页面内直接使用 axios。
- 方法命名：`getXxxList`、`createXxx`、`updateXxx`、`deleteXxx`，返回类型明确（`Promise<PageResult<UserItem>>`）。
- 请求参数与响应类型写在 `src/types/<module>.ts`，通过 `import type` 引入。

```ts
export const getUserList = (params?: Partial<UserListReq>) => {
  return request<PageResult<UserItem>>({ url: '/admin/user/list', method: 'get', params })
}
```

### 3.3 类型规范

- 复用公共类型：`ApiResponse`、`PageQuery`、`PageResult`、`TreeNode<T>`、`BaseEntity`。
- 禁止 `any`；不确定类型用 `unknown` 并在使用前收窄。
- 列表查询参数继承 `PageQuery`，分页响应使用 `PageResult<T>`。
- 后端树结构为 `{ data, children }`，`el-tree-select` 等组件需要平铺时在项目内做转换，并写注释说明原因。

### 3.4 语法与风格

- 纯 ES6+，函数统一使用**箭头函数**，不使用 `function` 声明。
- 组件使用 `<script setup lang="ts">`，组合式 API；`ref` / `reactive` 显式标注类型。
- 异步流程使用 `async/await`，加载态放在 `finally` 中关闭。
- 枚举集中在 `src/enums`，使用 `as const` 对象 + 文案映射，数值与后端一致（从 1 开始）。

### 3.5 页面与样式

- 优先复用 `src/styles/index.css` 中的颜色、间距变量与公共类：`.page-card`、`.toolbar`、`.table-operations`。
- 明暗主题必须同时可用，禁止硬编码只适配白底的颜色。
- 适配常用桌面宽度，窄屏要有合理降级。
- 所有日期时间通过 `src/utils/datetime.ts` 的 `formatDateTime` / `formatDateTimeCell` 格式化到秒，不展示毫秒。

### 3.6 权限与上传

- 按钮级权限使用指令：`v-perm="'POST:/admin/user/add'"`。
- 上传使用通用上传组件；提交业务数据时传 `relative_path`，不要把访问域名写进业务字段。
- 富文本编辑统一复用管理端 `src/components/RichTextEditor.vue`，不得在业务模块或插件中重复封装 WangEditor；上传、明暗主题和销毁生命周期由通用组件维护。

### 3.7 提交前检查

```bash
pnpm exec vue-tsc --noEmit
pnpm build
```

同时确认：亮/暗主题正常、无权限角色看不到按钮且后端返回 403、空态与错误态处理正确、未提交 `.env.local`/构建产物/密钥。

---

## 四、通用协作规范

1. **分支与提交**：功能分支开发，提交信息使用 `feat/fix/docs/refactor/chore: 描述` 前缀，一次提交只做一件事。
2. **敏感文件**：`config.yaml`、`.env*`、证书、密钥、构建产物一律不入库；模板文件（`config.example.yaml`、`.env.example`）必须保留并同步更新。
3. **接口文档**：后端接口变更同步 `docs/admin_openapi.yaml`，并更新 Apifox 在线文档与 README 顶部文档地址。
4. **双端一致**：字段命名、枚举取值、分页结构（`list` + `total`）、响应结构（`code`/`msg`/`data`）前后端必须一致。
5. **不要过度设计**：新增能力优先复用现有 `pkg/` 与 `common/`；只有在确认无复用可能时才新增包。
6. **发版说明**：每次发版必须更新 `config.example.yaml` 的 `version` 配置（版本号唯一来源是 `config.yaml`，服务启动时校验必填），并在 `docs/update_doc/` 新增 `v{version}.md` 版本功能说明，内容参照 `v0.0.1.md` / `v0.0.2.md` 的结构（版本目标、主要变更、数据库升级、部署顺序、验证命令）。
