---
name: go-frame-code-standards
description: This skill should be used when writing, modifying, or reviewing code in the go_backend_frame repository (Go backend `server_api` and Vue 3 admin frontend `admin_client`). It enforces the project's directory structure, layering (controller/logic/param/resp), unified response and permission conventions, and the pre-submit checklists for both Go and TypeScript.
---

# go_backend_frame 代码规范

## 目的

在 `go_backend_frame`（后端 `server_api` + 前端 `admin_client`）中新增或修改代码时，保证目录结构、分层方式、命名风格、响应格式与权限边界与现有工程一致。

## 使用场景

- 新增/修改后端接口（controller、logic、param、resp、model、路由、中间件）。
- 新增/修改前端页面、接口封装、类型、枚举、组件。
- 评审代码是否符合本项目约定。
- 新增数据库表或配置字段时确定脚本与配置落点。

## 工作流程

### 1. 确定改动端与落点

- 后端改动先判断属于管理端（`internal/admin`）还是用户端（`internal/api`），两端共用的能力放 `internal/common/<功能>`，无业务归属的工具放 `pkg/<功能>`。
- 前端改动按模块定位到 `src/api/<module>.ts`、`src/types/<module>.ts`、`src/views/<模块>/<页面>/index.vue`。

### 2. 后端改动按分层落地（顺序）

1. `internal/common/model/` 定义或复用 GORM 模型（表结构脚本放 `sql/`）。
2. `internal/<端>/param/` 新增请求结构体（按功能分文件，字段带 `json`/`form`/`binding`/`comment`）。
3. `internal/<端>/resp/` 新增响应结构体（列表用 `list` + `total`；树用 `data` + `children`）。
4. `internal/<端>/logic/` 写业务规则，方法统一接收 `*gin.Context`。
5. `internal/<端>/controller/` 只做绑定与响应。
6. `router/admin.go` 或 `router/api.go` 注册路由，挂到正确的中间件组。

详细模板与禁止事项见 `references/backend-go.md`。

### 3. 前端改动按模块落地

1. `src/types/<module>.ts` 定义请求与响应类型（复用 `BaseEntity`、`PageQuery`、`PageResult`、`TreeNode`）。
2. `src/api/<module>.ts` 用 `request<T>()` 封装，禁止直接使用 axios。
3. `src/enums/<module>.ts` 定义枚举（数值与后端一致，从 1 开始）。
4. `src/views/<模块>/<页面>/index.vue` 使用 `<script setup lang="ts">` 与箭头函数。

详细约定见 `references/frontend-vue.md`。

### 4. 收尾检查

- 后端：`gofmt -l .` 无输出、`go vet ./...`、`go build ./...` 通过。
- 前端：`pnpm exec vue-tsc --noEmit`、`pnpm build` 通过。
- 接口变更同步 `docs/admin_openapi.yaml` 与 Apifox 文档。
- 配置新增同步 `config/config.go`、`config.example.yaml` 与 README 配置表。
- 未提交 `config.yaml`、`.env*`、证书、密钥、构建产物。

## 速查规则

| 主题 | 规则 |
| --- | --- |
| 响应 | 一律 `response.OK/Fail`，HTTP 200，业务码 0/400/401/403/500/503 |
| 上下文 | controller 必须把 `*gin.Context` 传入 logic |
| 参数 | 请求结构体放 `param` 且按功能分文件，禁止合并为 `param.go` |
| 权限 | 按钮权限格式 `METHOD:/path`；免鉴权接口登记 `enums.SkipPermissionApis` |
| 缓存 | 菜单/角色/用户权限变更后必须清 `perm:<user_id>` 或全量权限缓存 |
| 密钥 | 响应结构不返回私钥/公钥/Secret；配置模板用占位符 |
| 上传 | 业务只存 `relative_path`，地址按默认存储动态生成，单文件 ≤50MB |
| 枚举 | 集中在 `common/enums`，数值从 1 开始，前端 `src/enums` 保持一致 |
| 数据库 | 禁止 AutoMigrate，变更提供 `sql/` 脚本 |
| 前端请求 | 统一 `request<T>()`，`baseURL` 来自 `VITE_ADMIN_API_BASE_URL` |
| 前端时间 | 用 `formatDateTime` 格式化到秒 |
| 前端权限 | 按钮用 `v-perm="'POST:/admin/user/add'"` |

## 参考文件

- `../CODE_STYLE.md`：完整规范说明（人工阅读版）。
- `references/backend-go.md`：后端目录结构、分层职责、代码模板、常见禁止事项。
- `references/frontend-vue.md`：前端目录结构、接口/类型/页面模板、样式与提交检查。
