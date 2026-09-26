# 插件开发、分发、安装与更新指南

## 1. 方案说明

从核心 `v0.0.4` 开始，新插件使用“独立插件仓库 + 标准 ZIP 发行包 + 安装器 CLI + 编译期注册”模式。

开发者不再把插件开发分支反复合并到核心主分支，安装人员也不再手工修改 `internal/plugins/register.go`。安装器负责：

- 校验 ZIP 路径穿越、符号链接、包大小和文件数量；
- 校验 `plugin.json`、核心版本、迁移 SHA256、菜单业务键；
- 限制迁移 SQL 只能操作当前插件的 `plg_{plugin_id}_*` 表；
- 安装或升级后端插件目录；
- 同时安装或升级管理端插件目录；
- 保存数据库迁移文件；
- 自动扫描插件并生成 `register_gen.go`；
- 使用菜单业务键创建数据库自增 ID，避免插件指定菜单 ID；
- 可选执行数据库迁移并写入插件、迁移和安装日志；
- 源码安装失败时恢复原插件目录并重新生成注册文件。

当前仍是源码级插件。安装后必须重新检查、构建并发布后端和管理端，不支持运行时加载未知 Go 代码或远程 JavaScript。

## 2. Git 与版本基线

### 2.1 核心版本

接口端和管理端必须使用同一核心标签。正式开发新插件时统一基于 `v0.0.4`：

```bash
git fetch --tags --prune
git switch -c plugin/news-v1.0.0 v0.0.4
```

在 `v0.0.4` 标签尚未发布时，只能临时从 `release/v0.0.4` 开发；标签发布后必须重新对齐标签并完成全部检查，不得继续从 `v0.0.2` 或其他旧核心版本创建新插件。

核心维护者发布标签：

```bash
git switch main
git pull --ff-only
git tag -a v0.0.4 -m "release v0.0.4"
git push origin v0.0.4
```

### 2.2 插件独立仓库

推荐一个插件一个仓库：

```text
news-plugin/
├── plugin.json
├── README.md
├── docs/
│   ├── API.md
│   ├── DATABASE.md
│   ├── api/
│   │   └── v1.0.0.md
│   ├── database/
│   │   └── v1.0.0.md
│   └── updates/
│       └── v1.0.0.md
├── CHANGELOG.md
├── LICENSE
├── server_api/
├── admin_client/
└── database/
```

插件仓库使用自己的分支和标签：

```text
main
feature/article-search
fix/publish-race
release/v1.1.0

v1.0.0
v1.1.0
```

首次开发从核心 `v0.0.4` 模板创建插件仓库。后续更新从插件仓库自己的 `main` 或上一个插件标签开发，不需要再次合并开发商的历史分支到核心仓库；升级已有插件时，应先建立兼容 `v0.0.4` 的发行版本，再继续增加新功能。

## 3. 标准发行包结构

```text
news-v1.0.0/
├── plugin.json
├── README.md
├── docs/
│   ├── API.md
│   ├── DATABASE.md
│   ├── api/
│   │   └── v1.0.0.md
│   ├── database/
│   │   └── v1.0.0.md
│   └── updates/
│       └── v1.0.0.md
├── CHANGELOG.md
├── LICENSE
├── server_api/
│   └── internal/plugins/news/
│       ├── plugin.go
│       ├── model/
│       ├── enums/
│       ├── service/
│       ├── admin/
│       └── api/
├── admin_client/
│   └── src/plugins/news/
│       ├── api/
│       ├── types/
│       ├── enums/
│       ├── components/
│       └── views/
└── database/
    └── migrations/
        └── v1.0.0.sql
```

固定规则：

- 后端入口必须位于 `server_api/internal/plugins/{plugin_id}/plugin.go`；
- 后端插件必须导出 `New()` 并实现 `commonplugin.Plugin`；
- 管理端插件只能位于 `admin_client/src/plugins/{plugin_id}`；
- 迁移文件必须位于发行包 `database/` 目录；
- 安装器不会接收插件提供的完整 `register.go`；
- 安装器不会执行插件提供的菜单 INSERT SQL；
- 插件包不得包含配置文件、密钥、证书、构建缓存、`node_modules` 或 `dist`。
- 插件包必须包含非空且可离线阅读的 `docs/api/v{version}.md`，并覆盖发行版本提供的全部管理端、用户端和回调接口；安装器按 `plugin.json.version` 强制检查该文件。
- 插件包必须包含非空且可离线阅读的 `docs/database/v{version}.md`，完整说明业务表、内外部引用、文件归属以及 Remove/Purge 影响；没有引用也必须明确写“无”。安装器会强制检查并归档该文件。
- 数据库说明必须明确包含是否引用核心业务表、是否引用其他插件表、是否允许 Purge、是否自动删除文件等结论；只创建空文件或省略结论不能通过校验。

`logo` 是当前存储渠道中的相对路径，不是插件包内静态文件路径。如果安装环境没有预先上传对应文件，应在清单中留空，安装后通过插件管理页面上传 Logo。

## 4. plugin.json

完整示例：

```json
{
  "plugin_id": "news",
  "name": "新闻资讯",
  "version": "1.0.0",
  "core_version": "0.0.4",
  "logo": "plugins/news/logo.png",
  "author": "开发者或维护团队",
  "homepage": "https://example.com/plugins/news",
  "description": "提供新闻分类、文章发布和用户端阅读能力",
  "dependencies": [],
  "migrations": [
    {
      "version": "1.0.0",
      "file": "database/migrations/v1.0.0.sql",
      "checksum": "迁移文件的64位SHA256",
      "description": "创建新闻资讯基础表"
    }
  ],
  "menus": [
    {
      "key": "news",
      "parent_key": "",
      "name": "新闻资讯",
      "type": 1,
      "path": "/plugin/news",
      "api_path": "",
      "icon": "Document",
      "sort": 10,
      "status": 1,
      "remark": "新闻资讯插件目录"
    },
    {
      "key": "article",
      "parent_key": "news",
      "name": "文章管理",
      "type": 2,
      "path": "/plugin/news/article",
      "api_path": "GET:/admin/plugin/news/article/list",
      "icon": "DocumentCopy",
      "sort": 1,
      "status": 1,
      "remark": "新闻文章管理"
    },
    {
      "key": "article_create",
      "parent_key": "article",
      "name": "新增文章",
      "type": 3,
      "path": "",
      "api_path": "POST:/admin/plugin/news/article/create",
      "icon": "",
      "sort": 1,
      "status": 1,
      "remark": "新增新闻文章"
    }
  ]
}
```

### 4.1 插件 ID

- 长度 2～64 位；
- 以小写字母开头；
- 只能包含小写字母、数字和下划线；
- 发布后不得修改；
- 后端目录、前端目录、表前缀、菜单路径和迁移记录必须一致。

### 4.2 菜单业务键

菜单使用 `key` 和 `parent_key` 描述层级，禁止提供 `id`、`parent_id`：

```text
news
└── article
    └── article_create
```

安装器按父级顺序创建菜单，由数据库生成真实 `sys_menu.id`，然后写入 `sys_plugin_menu` 映射表。

`plugin_id + menu_key` 是唯一业务键。升级时安装器通过该业务键更新原菜单，不依赖不同数据库中的数字 ID。

插件清单中的默认层级只能引用同一插件包中的 `parent_key`，不允许写入某个环境的核心菜单数字 ID。安装后，管理员可以在菜单管理中把插件目录或页面移动到核心菜单或其他目录下面；系统会把该映射标记为管理员自定义，后续插件升级保留实际 `parent_id`，不再用清单默认父级覆盖。插件按钮必须保留在所属页面下，不能单独移动。

如果管理员从未调整父级，升级器仍按新版本清单同步默认层级。无论父级是否自定义，菜单名称、类型、路径、接口、图标、排序、状态和备注继续由插件清单同步。

## 5. 开发插件

### 5.1 后端

后端仍遵循 [插件开发规范](./plugin-development-standard.md)。`plugin.go` 示例：

```go
func (p *Plugin) Manifest() commonplugin.Manifest {
    return commonplugin.Manifest{
        PluginID:    "news",
        Name:        "新闻资讯",
        Version:     "1.0.0",
        Logo:        "plugins/news/logo.png",
        Author:      "开发者或维护团队",
        Homepage:    "https://example.com/plugins/news",
        Description: "提供新闻分类、文章发布和用户端阅读能力",
        CoreVersion: commonplugin.CoreVersion,
    }
}
```

开发者不再修改 `internal/plugins/register.go`。安装器会执行与下面等价的命令：

```bash
go run . plugin generate --server-root .
```

生成文件：

```text
internal/plugins/register_gen.go
```

该文件带有 `Code generated` 标记，不允许手工编辑。

### 5.2 管理端

页面路径：

```text
/plugin/news/article
```

对应：

```text
admin_client/src/plugins/news/views/article/index.vue
```

管理端插件目录与后端放在同一个发行包中。执行安装或升级命令时通过 `--admin-root` 一起安装，不需要再次复制前端代码。

## 6. 数据库迁移安全

### 6.1 允许操作

迁移 SQL 仅允许操作当前插件表：

```text
plg_{plugin_id}_*
```

允许的语句类型：

- `CREATE TABLE`；
- `ALTER TABLE`；
- `CREATE INDEX`；
- `INSERT INTO` 当前插件表；

`INSERT` 必须显式声明字段列表，且不能包含自增 `id`。

### 6.2 禁止操作

安装器拒绝：

- `DROP`、`TRUNCATE`、`DELETE`、`REPLACE`、`RENAME`；
- 任意业务数据 `UPDATE`、`ON DUPLICATE KEY UPDATE`；
- 指定自增 `id` 的 INSERT；
- 删除、改名或修改现有字段的破坏性 `ALTER TABLE`；
- 存储过程、函数、触发器和事件；
- `LOAD DATA`、`OUTFILE`、授权语句；
- 任何包含 `sys_` 核心表的迁移；
- 操作其他插件表；
- 插件自行写入 `sys_menu`、`sys_role_menu`、`sys_plugin`；
- 通过 SQL 指定菜单 ID 和父级 ID。

菜单、迁移记录和插件信息由安装器写入。

### 6.3 迁移摘要

macOS：

```bash
shasum -a 256 database/migrations/v1.0.0.sql
```

Linux：

```bash
sha256sum database/migrations/v1.0.0.sql
```

将摘要写入 `plugin.json` 和后端 `Migrations()`。修改 SQL 后必须重新生成摘要。已发布的迁移文件不得修改，只能新增更高版本迁移。

MySQL DDL 可能隐式提交，无法保证所有 DDL 完全事务回滚。因此迁移仍需可重复执行，并必须先在测试数据库验证。安装器的安全检查降低误操作范围，但不能替代备份。

## 7. 校验和打包

`plugin.json` 必须位于 ZIP 根目录。进入插件发行目录后执行：

```bash
cd news-v1.0.0
zip -r ../news-v1.0.0.zip .
cd ..
shasum -a 256 news-v1.0.0.zip
```

使用核心 CLI 校验：

```bash
cd server_api
go run . plugin validate /path/news-v1.0.0.zip
```

校验内容包括：

- ZIP 最大 100MB；
- 文件数量不超过 2000；
- 禁止绝对路径、`../` 路径穿越和符号链接；
- 核心版本必须精确匹配；
- 后端入口必须存在；
- 菜单业务键唯一且父级存在；
- 页面菜单路径必须位于 `/plugin/{plugin_id}/`；
- 迁移文件 SHA256 一致；
- 迁移 SQL 只能操作当前插件表。

## 8. 安装插件

### 8.1 安装前

1. 验证发行包 SHA256；
2. 审查源码、SQL、依赖和许可证；
3. 备份数据库、后端和管理端构建产物；
4. 确认后端和管理端均为兼容核心版本；
5. 在测试环境先安装；
6. 确保工作区没有会被插件目标目录覆盖的未提交内容。

### 8.2 只安装源码

```bash
cd server_api
go run . plugin install /path/news-v1.0.0.zip \
  --server-root . \
  --admin-root ../admin_client
```

该命令会：

- 安装 `internal/plugins/news`；
- 安装 `admin_client/src/plugins/news`；
- 保存迁移文件到 `sql/plugins/news/1.0.0`；
- 归档版本接口、数据库和更新说明到 `plugin_docs/news/1.0.0`；
- 自动生成 `internal/plugins/register_gen.go`；
- 不连接数据库。

只安装源码后，数据库中不会出现插件记录、迁移记录、菜单映射或安装日志，命令输出会显示 `数据库=false`。如需一次完成完整安装，应直接使用下一节的 `--apply-database` 命令。

如果已经完成源码安装、之后才准备好数据库，应继续使用同一个发行包执行 `plugin install --apply-database` 补齐数据库。安装器需要支持该补录场景：重新校验发行包，原子替换包内源码，然后执行迁移、菜单、插件记录和安装日志。此时不能使用 `plugin upgrade`，因为数据库中尚不存在已安装插件记录；也不应要求使用者手工删除插件目录。

### 8.3 同时应用数据库

```bash
go run . plugin install /path/news-v1.0.0.zip \
  --server-root . \
  --admin-root ../admin_client \
  --config config.yaml \
  --apply-database
```

数据库安装器会：

- 为当前插件获取数据库级互斥锁，同一插件的安装或升级只能串行执行；
- 先执行可安全重试的业务表迁移；
- 在同一个事务中写入迁移记录、系统菜单、菜单映射、插件信息和成功安装日志；
- 任一元数据写入失败时回滚整个元数据事务，不留下孤立菜单；
- 失败后可使用同一发行包重试，不重复创建已经成功登记的菜单。

MySQL 的 DDL 会隐式提交，不能和菜单等 DML 实现真正的单事务回滚。因此迁移脚本必须使用可重复执行的建表或加字段方案；安装器保证的是插件元数据原子提交，并通过迁移幂等性保证 DDL 阶段可以安全重试。

新插件安装后默认写入 `status=2`（停用）。安装完成并不代表插件路由已经注册：必须先在插件管理页面启用，再重启管理端 API 和用户端 API。启用前访问 `/admin/plugin/{plugin_id}/*` 或 `/api/plugin/{plugin_id}/*` 返回 404 是正常行为；权限不足则应返回 403，而不是 404。

1. 检查插件基础表；
2. 检查插件是否已经安装；
3. 检查依赖插件是否启用且版本一致；
4. 再次校验迁移摘要；
5. 执行未登记的安全迁移；
6. 写入 `sys_plugin_migration`；
7. 按业务键创建菜单和 `sys_plugin_menu` 映射；
8. 以停用状态写入 `sys_plugin`；
9. 写入 `sys_plugin_install_log`。

安装器不会自动启用插件。

### 8.4 构建和启用

接口端：

```bash
gofmt -l .
go vet ./...
go test ./...
go build ./...
```

管理端：

```bash
pnpm exec vue-tsc --noEmit
pnpm build
```

检查通过后：

1. 发布后端和管理端构建产物；
2. 在“系统维护 → 插件管理”确认安装版本和程序版本一致；
3. 启用插件；
4. 重启管理端 API 和用户端 API；
5. 给普通角色分配插件菜单和按钮权限；
6. 完成冒烟测试。

## 9. 更新插件

### 9.1 版本升级规则

插件版本必须遵循 SemVer，固定使用 `主版本.次版本.修订版本`：

| 变化类型 | 升级方式 | 示例 |
| --- | --- | --- |
| 存在不兼容的接口、配置或数据结构变化 | 提升主版本，次版本和修订版本归零 | `1.4.2 → 2.0.0` |
| 新增向后兼容的业务能力 | 提升次版本，修订版本归零 | `1.4.2 → 1.5.0` |
| 向后兼容的问题修复、UI优化、文档或静态资源变化 | 提升修订版本 | `1.4.2 → 1.4.3` |

每一次发布都必须使用高于已安装版本的新版本号。禁止修改并重新分发同版本 ZIP，即使变化仅涉及前端样式或文档。`plugin.json.version`、后端 `Manifest.Version`、Git Tag 和 ZIP 文件名必须一致。升级器会拒绝相同版本或更低版本。

每个发行版本必须新增 `docs/updates/v{version}.md`，内容至少包括：

- 版本类型和适用范围；
- 新增、调整和修复内容；
- 数据库迁移及数据兼容影响，没有变化时明确写“无”；
- 安装或升级命令以及构建、重启要求；
- 兼容性、风险和回滚方法。

安装器会校验当前版本对应的更新说明存在且非空，缺少时拒绝校验、安装和升级。

### 9.2 开发商流程

插件开发商从插件仓库自己的稳定版本开发：

```bash
git switch main
git pull --ff-only
git switch -c feature/article-search
```

完成后合并回插件仓库 `main`，新增迁移文件，更新：

- `plugin.json.version`；
- 后端 `Manifest.Version`；
- `plugin.json.migrations`；
- `plugin.json.menus`；
- `CHANGELOG.md`；
- `docs/api/v{version}.md`；
- `docs/database/v{version}.md`；
- `docs/updates/v{version}.md`。

然后发布新标签和新 ZIP：

```bash
git tag -a v1.1.0 -m "release news plugin v1.1.0"
```

不需要把旧插件开发分支重新合并到核心主分支。

### 9.3 使用方升级

升级前必须先在插件管理页面停用插件并重启服务。然后执行：

```bash
cd server_api
go run . plugin upgrade /path/news-v1.1.0.zip \
  --server-root . \
  --admin-root ../admin_client \
  --config config.yaml \
  --apply-database
```

升级器要求：

- 插件已经安装；
- 插件当前处于停用状态；
- 目标版本高于当前版本；
- 依赖插件已启用且版本一致；
- 已执行迁移的摘要不得改变。

升级时通过菜单业务键更新原菜单，不创建固定 ID。新菜单创建新的自增 ID；未出现在新清单中的旧菜单不会自动删除，避免升级误删角色权限，需由明确的弃用迁移处理。

升级完成后重新构建、发布、启用并重启两个 API 服务。

## 10. 失败与恢复

- ZIP 校验失败时不会写入源码或数据库；
- 源码复制或注册代码生成失败时恢复原插件目录；
- 数据库处理失败时插件保持停用，不会被运行时加载；
- MySQL DDL 可能已经部分提交，应根据错误检查迁移并从备份恢复；
- 不允许通过修改迁移摘要绕过校验；
- 不允许直接把数据库插件状态改成启用来绕过安装器。

安装或升级完成后如果编译失败，修复插件源码或恢复安装前代码，再执行：

```bash
go run . plugin generate --server-root .
```

## 11. 停用和卸载

停用：

1. 在插件管理页面停用；
2. 重启管理端 API 和用户端 API；
3. 确认插件路由和后台任务不再运行。

停用不会删除业务表、菜单、授权、文件、迁移记录或源码。

可恢复移除：

```bash
go run . plugin remove news \
  --server-root . \
  --admin-root ../admin_client \
  --config config.yaml
```

执行条件和结果：

1. 插件必须已经停用并完成服务重启；
2. 不能存在依赖该插件的已启用插件；
3. 在插件级数据库锁内，用单个事务删除角色授权、菜单、菜单映射和插件记录；
4. 保留业务表、迁移历史、安装日志和上传文件；
5. 删除前后端源码并重新生成 `register_gen.go`；
6. 命令中断后可以重复执行，数据库记录已经移除时仍会继续清理源码。

彻底清理：

```bash
go run . plugin purge news \
  --server-root . \
  --admin-root ../admin_client \
  --config config.yaml \
  --confirm news
```

Purge 不可恢复，必须先备份。命令首先要求宿主最新归档版本中存在非空的 `database/v{version}.md`，再完成可恢复移除并删除 `plg_news_*` 业务表；所有业务表删除完成后才事务删除迁移历史和安装日志。若 DDL 阶段中断，插件已不可运行，重新执行相同命令会从剩余业务表继续。检测到插件表被其他前缀的表通过外键引用时会拒绝清理。上传文件不会自动删除，需要根据数据库说明和业务归属单独审查。

## 12. 发行检查清单

- [ ] 接口端和管理端使用同一核心版本。
- [ ] 插件 ID、版本、核心版本和依赖正确。
- [ ] 后端 `New()` 和 Manifest 正确。
- [ ] 管理端页面位于标准插件目录。
- [ ] 列表页与新增/修改表单组件已拆分，表单交互没有堆放在列表页中。
- [ ] 菜单只使用 `key`、`parent_key`，没有数据库 ID。
- [ ] 页面路径位于 `/plugin/{plugin_id}/`。
- [ ] 迁移只操作 `plg_{plugin_id}_*` 表。
- [ ] 迁移没有 `DROP`、`DELETE`、核心表或其他插件表操作。
- [ ] 迁移 SHA256 与清单及代码一致。
- [ ] ZIP 不包含配置、密钥、证书和构建缓存。
- [ ] `docs/api/v{version}.md` 已包含在仓库和 ZIP 中，覆盖全部实际路由、鉴权、参数、响应、错误场景及版本变更。
- [ ] `docs/database/v{version}.md` 已包含在仓库和 ZIP 中，完整声明业务表、内外部数据库引用、文件归属和移除影响；无引用项已明确写“无”。
- [ ] `plugin validate` 通过。
- [ ] 后端测试、Vet、构建通过。
- [ ] 管理端类型检查、生产构建通过。
- [ ] 全新安装和上一版本升级均在测试数据库验证。

## 13. 安装检查清单

- [ ] 已验证 ZIP 摘要、源码、SQL、依赖和许可证。
- [ ] 已备份数据库和构建产物。
- [ ] 已确认插件目标目录没有未提交修改。
- [ ] 已先在测试环境执行安装或升级。
- [ ] 已检查 `register_gen.go` 只包含预期插件。
- [ ] 已检查迁移、菜单映射和安装日志。
- [ ] 已完成后端和管理端构建。
- [ ] 已发布两端构建产物。
- [ ] 已启用并重启两个 API 服务。
- [ ] 已分配角色权限并完成冒烟测试。

## 14. 常见问题

### 为什么还需要重新构建？

Go 插件采用编译期静态注册。安装器消除了手工修改注册文件，但不能绕过 Go 编译过程。

### 为什么不自动修改 register.go？

`register.go` 保持稳定，只调用自动生成的 `generatedBuiltins()`。安装器只维护 `register_gen.go`，多个插件不会再争抢同一个人工文件。

### 为什么插件 SQL 不能写菜单？

不同数据库的自增 ID 不一致。菜单由 `plugin.json` 的业务键描述，安装器统一生成 ID 和父子关系，避免覆盖数据或挂错层级。

### 为什么更新不再合并旧分支？

插件拥有独立仓库和版本。开发商发布新 ZIP，使用方运行 `plugin upgrade`，只替换当前插件的后端、管理端和版本迁移目录。

### 为什么默认不执行数据库？

数据库变更风险高，必须显式提供 `--apply-database`。这样代码审查和数据库审批可以分阶段完成。

### 为什么没有自动卸载？

卸载通常包含删除表、菜单、授权和文件，是不可逆操作。当前版本优先保证数据安全，只支持停用和人工审查卸载。
