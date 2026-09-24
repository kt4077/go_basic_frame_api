# 插件开发、分发、安装与更新指南

## 1. 文档目的

本文面向第一次参与插件开发的开发者，说明如何从核心框架 `v0.0.2` 创建插件开发分支，如何同时组织后端、管理端和 SQL，如何生成可审核的插件发行包，以及使用方如何安装、启用、升级、停用和卸载插件。

编码细则以 [插件开发规范](./plugin-development-standard.md) 为准。本文重点描述 Git、版本、交付和部署流程。

## 2. 当前插件形态

`v0.0.2` 使用源码级、编译期插件：

- 插件后端源码编译进 `server_api`；
- 插件管理端源码编译进 `admin_client`；
- 插件数据库变更通过显式 SQL 执行；
- 插件必须在 `internal/plugins/register.go` 显式注册；
- 插件启停在管理端维护，但重启管理端 API 和用户端 API 后才生效；
- 不支持上传 ZIP 后直接执行未知 Go 代码；
- 不支持 Go `.so` 动态库、远程 JavaScript 或运行时热卸载；
- 第三方插件安装前必须完成源码和 SQL 审查。

因此，“分发插件”是分发可审查的源码发行包，不是分发一个可在页面直接运行的二进制扩展。

## 3. 仓库和版本基线

插件通常同时涉及两个仓库：

| 仓库 | 内容 | 地址 |
| --- | --- | --- |
| 接口端 | Go 插件、迁移 SQL、接口和生命周期 | [go_basic_frame_api](https://gitee.com/open-source-project-open/go_basic_frame_api) |
| 管理端 | Vue 插件页面、接口封装、类型和枚举 | [go_basic_frame_admin](https://gitee.com/open-source-project-open/go_basic_frame_admin) |

核心版本使用 Git Tag 作为插件开发的唯一稳定基线。分支会继续变化，不能作为已发布插件的兼容性凭据。

### 3.1 核心仓库分支约定

| 类型 | 命名 | 用途 |
| --- | --- | --- |
| 稳定主分支 | `main` | 已合并并准备发布的核心代码 |
| 发布准备分支 | `release/v0.0.2` | `v0.0.2` 发布前的短期稳定分支 |
| 核心功能分支 | `feature/<date>_<feature>` | 核心框架能力开发 |
| 修复分支 | `fix/<issue-or-summary>` | 核心缺陷修复 |
| 发布标签 | `v0.0.2` | 插件开发和兼容性声明的固定基线 |

### 3.2 插件分支约定

插件开发分支统一使用：

```text
plugin/<plugin_id>-v<plugin_version>
```

示例：

```text
plugin/news-v1.0.0
plugin/shop-v1.0.0
plugin/shop-v1.1.0
```

接口端和管理端建议使用相同分支名，便于确认两个仓库属于同一次插件发行。

### 3.3 关于当前 v0.0.2 标签

在本文编写时，仓库尚未正式创建 `v0.0.2` 标签。核心维护者发布前应在接口端和管理端分别完成：

```bash
git switch main
git pull --ff-only
git tag -a v0.0.2 -m "release v0.0.2"
git push origin v0.0.2
```

两个仓库的 `v0.0.2` 标签应代表同一套接口和页面契约。

标签发布前如必须并行开发，只能临时从 `release/v0.0.2` 创建插件分支。正式标签发布后，插件开发者必须将插件分支变基或重新移植到 `v0.0.2`，并重新完成全部检查。

不要从个人临时分支或未发布的 `feature/*` 分支对外发布插件。

## 4. 新开发者开始开发

以下以 `news` 插件 `1.0.0` 为例。

### 4.1 准备环境

- Go 版本与接口端 `go.mod` 一致；
- Node.js、pnpm 版本满足管理端 README；
- MySQL、Redis 已启动；
- 已准备独立开发数据库；
- 已配置接口端 `config.yaml` 和管理端 `.env.development`；
- 不使用生产数据库开发插件。

### 4.2 拉取接口端并从标签建分支

```bash
git clone https://gitee.com/open-source-project-open/go_basic_frame_api.git
cd go_basic_frame_api
git fetch --tags --prune
git switch -c plugin/news-v1.0.0 v0.0.2
```

确认基线：

```bash
git describe --tags --exact-match v0.0.2
git status
```

### 4.3 拉取管理端并从标签建分支

```bash
git clone https://gitee.com/open-source-project-open/go_basic_frame_admin.git
cd go_basic_frame_admin
git fetch --tags --prune
git switch -c plugin/news-v1.0.0 v0.0.2
```

不要在一个仓库使用 `main`，另一个仓库使用 `v0.0.2`。两端必须使用相同核心版本基线。

### 4.4 选择插件标识

发布前确定全局唯一的 `plugin_id`：

```text
news
shop
content_center
```

规则：

- 长度 2～64 位；
- 以小写字母开头；
- 只能包含小写字母、数字和下划线；
- 发布后不得修改；
- 后端目录、前端目录、表前缀、菜单路径和迁移记录必须使用同一个 ID。

## 5. 创建插件代码

### 5.1 后端目录

在接口端创建：

```text
internal/plugins/news/
├── plugin.go
├── enums/
├── model/
├── service/
├── admin/
│   ├── controller/
│   ├── logic/
│   ├── param/
│   └── resp/
└── api/
    ├── controller/
    ├── logic/
    ├── param/
    └── resp/
```

SQL 放在：

```text
sql/plugins/news/v1.0.0/
├── migration.sql
├── register.sql
└── uninstall.sql
```

文件职责：

| 文件 | 作用 |
| --- | --- |
| `migration.sql` | 建表、索引、基础数据、菜单和权限，不写迁移成功记录 |
| `register.sql` | 写入 `sys_plugin_migration` 和 `sys_plugin`，包含 migration.sql 的 SHA256 |
| `uninstall.sql` | 人工卸载参考，默认不自动删除数据 |

### 5.2 管理端目录

在管理端创建：

```text
src/plugins/news/
├── api/
├── components/
├── enums/
├── types/
└── views/
    ├── article/
    │   └── index.vue
    └── category/
        └── index.vue
```

菜单路径与页面对应：

```text
/plugin/news/article
→ src/plugins/news/views/article/index.vue
```

### 5.3 实现 Manifest

`plugin.go` 至少包含：

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
        Dependencies: []commonplugin.Dependency{},
    }
}
```

`CoreVersion` 必须使用 `commonplugin.CoreVersion`。插件 `1.0.0` 与核心 `0.0.2` 是两个不同版本，不要混淆。

### 5.4 注册插件

修改 `internal/plugins/register.go`：

```go
import (
    commonplugin "server_api/internal/common/plugin"
    "server_api/internal/plugins/news"
)

func RegisterBuiltins(registry *commonplugin.Registry) error {
    builtins := []commonplugin.Plugin{
        news.New(),
    }
    // 保留现有注册循环
}
```

每个发行包都必须明确说明对该文件的修改。如果多个插件同时安装，安装人员需要合并 `import` 和 `builtins`，不能用发行包直接覆盖整个文件。

## 6. 编写和登记数据库迁移

### 6.1 migration.sql

要求：

- 表名使用 `plg_news_` 前缀；
- 表和字段都有注释；
- 使用 `utf8mb4_general_ci`；
- 枚举从 1 开始；
- 包含合理的唯一索引和普通索引；
- 尽量可重复执行；
- 不使用 `AutoMigrate`；
- 不在服务启动时执行 DDL；
- 不写入 `sys_plugin_migration`，避免迁移文件摘要包含自身摘要形成循环。

### 6.2 生成 SHA256

macOS：

```bash
shasum -a 256 sql/plugins/news/v1.0.0/migration.sql
```

Linux：

```bash
sha256sum sql/plugins/news/v1.0.0/migration.sql
```

将结果同时写入：

1. `plugin.go` 的 `Migrations()`；
2. `register.sql` 的 `sys_plugin_migration.checksum`；
3. 发行包根目录 `checksums.txt`。

修改 `migration.sql` 后必须重新生成摘要。已正式发布的迁移文件禁止修改，只能新增更高版本迁移。

### 6.3 register.sql

`register.sql` 推荐使用事务并保持插件初始状态为停用：

```sql
START TRANSACTION;

INSERT INTO `sys_plugin_migration`
  (`created_at`, `updated_at`, `plugin_id`, `version`, `checksum`, `status`, `execution_ms`, `error_message`, `executed_at`)
VALUES
  (NOW(3), NOW(3), 'news', '1.0.0', 'migration.sql的SHA256', 1, 0, '', NOW(3))
ON DUPLICATE KEY UPDATE
  `plugin_id` = VALUES(`plugin_id`);

INSERT INTO `sys_plugin`
  (`created_at`, `updated_at`, `plugin_id`, `name`, `version`, `logo`, `author`, `homepage`, `description`, `status`, `manifest`)
VALUES
  (NOW(3), NOW(3), 'news', '新闻资讯', '1.0.0', 'plugins/news/logo.png', '开发者或维护团队',
   'https://example.com/plugins/news', '提供新闻分类、文章发布和用户端阅读能力', 2,
   '{"plugin_id":"news","name":"新闻资讯","version":"1.0.0","logo":"plugins/news/logo.png","core_version":"0.0.2"}')
ON DUPLICATE KEY UPDATE
  `plugin_id` = VALUES(`plugin_id`);

COMMIT;
```

安装脚本不得默认把插件设为启用。只有代码、前端和迁移全部发布并检查通过后，才可以在插件管理页面启用。

重复执行首次安装的 `register.sql` 不应覆盖已经存在的版本、摘要和启停状态。升级版本必须在新版本 `register.sql` 中校验旧版本后，再显式更新 `sys_plugin.version` 和清单快照。

## 7. 本地开发和联调

### 7.1 初始化核心数据库

```bash
mysql -uroot -p plugin_dev < sql/v0.0.1/cf_backend_frame.sql
mysql -uroot -p plugin_dev < sql/v0.0.2/plugin_base.sql
```

### 7.2 执行插件迁移

```bash
mysql -uroot -p plugin_dev < sql/plugins/news/v1.0.0/migration.sql
mysql -uroot -p plugin_dev < sql/plugins/news/v1.0.0/register.sql
```

### 7.3 构建并启动

接口端：

```bash
go mod download
go run . service admin -c config.yaml
go run . service api -c config.yaml
```

管理端：

```bash
pnpm install
pnpm dev
```

### 7.4 启用插件

1. 使用超级管理员登录管理端；
2. 打开“系统维护 → 插件管理”；
3. 确认安装版本和程序版本一致；
4. 查看迁移记录是否成功、摘要是否一致；
5. 点击启用；
6. 停止并重新启动管理端 API 和用户端 API；
7. 重新登录，检查插件菜单和按钮权限。

普通角色需要在角色管理中分配插件菜单和按钮权限。

### 7.5 开发阶段验证

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

还必须手工验证：

- 未登录、无权限和有权限三种访问状态；
- 管理端和用户端数据范围；
- 重复提交、并发写入和唯一索引冲突；
- 空数据、错误状态和分页边界；
- 亮色、暗色和常用桌面宽度；
- 插件启用、停用及重启行为；
- 依赖缺失、版本不一致和迁移摘要不一致时服务能拒绝启动；
- 卸载脚本不会误删核心或其他插件数据。

## 8. 插件版本和提交

### 8.1 插件版本

插件推荐使用语义化版本：

```text
主版本.次版本.修订版本
```

- `1.0.0`：首次稳定发布；
- `1.1.0`：向后兼容的新功能；
- `1.1.1`：向后兼容的缺陷修复；
- `2.0.0`：包含不兼容变更。

插件版本必须同时更新：

- `Manifest.Version`；
- `sys_plugin.version` 安装/升级 SQL；
- 发行包目录名；
- 发行说明；
- 管理端与接口端插件版本说明。

### 8.2 提交范围

一次插件发行应尽量只包含：

- `internal/plugins/<plugin_id>`；
- `internal/plugins/register.go` 的最小注册改动；
- `sql/plugins/<plugin_id>`；
- `src/plugins/<plugin_id>`；
- 必要的插件文档。

不要在插件分支顺带重构核心模块。确实需要核心能力时，先单独提交核心 Pull Request，等待进入新的核心版本，再更新插件的 `CoreVersion`。

### 8.3 提交信息建议

```text
feat(plugin-news): add article and category management
fix(plugin-news): prevent duplicate article publishing
docs(plugin-news): add installation guide
```

## 9. 生成插件发行包

### 9.1 推荐发行包结构

```text
news-v1.0.0-core-v0.0.2/
├── manifest.json
├── README.md
├── CHANGELOG.md
├── LICENSE
├── checksums.txt
├── server_api/
│   ├── internal/plugins/news/
│   ├── register.patch
│   └── sql/plugins/news/v1.0.0/
└── admin_client/
    └── src/plugins/news/
```

不要在发行包中包含：

- `config.yaml`、`.env.local`；
- 数据库密码、JWT 密钥、第三方密钥和证书；
- `node_modules`、`dist`、Go 构建缓存；
- 开发数据库、上传文件和日志；
- 无关核心代码；
- 未经授权的第三方资源。

### 9.2 manifest.json

发行包根清单示例：

```json
{
  "plugin_id": "news",
  "name": "新闻资讯",
  "version": "1.0.0",
  "core_version": "0.0.2",
  "logo": "plugins/news/logo.png",
  "author": "开发者或维护团队",
  "homepage": "https://example.com/plugins/news",
  "description": "提供新闻分类、文章发布和用户端阅读能力",
  "dependencies": [],
  "has_server": true,
  "has_admin_client": true
}
```

### 9.3 register.patch

不要直接分发完整的 `internal/plugins/register.go` 并要求覆盖。应生成只包含插件注册变化的补丁：

```bash
git diff v0.0.2 -- internal/plugins/register.go > register.patch
```

安装人员应人工检查和合并补丁。多个插件修改同一注册文件时必须解决冲突并保留所有插件。

### 9.4 发行包摘要

在发行目录外执行，摘要文件本身不参与计算，避免自引用：

```bash
find news-v1.0.0-core-v0.0.2 -type f ! -name checksums.txt -print0 \
  | sort -z \
  | xargs -0 shasum -a 256 \
  > news-v1.0.0-core-v0.0.2/checksums.txt
```

不同操作系统工具存在差异，也可以由 CI 生成摘要。发布页面应同时提供发行包本身的 SHA256，使用方下载后先验证再解压。

### 9.5 打包

```bash
zip -r news-v1.0.0-core-v0.0.2.zip news-v1.0.0-core-v0.0.2
shasum -a 256 news-v1.0.0-core-v0.0.2.zip
```

如分发闭源插件，也必须向安装方提供足够完成安全审查的源码。当前 `v0.0.2` 不支持只分发不可审查二进制插件。

## 10. 发布插件

推荐发布流程：

1. 接口端和管理端插件分支全部完成检查；
2. 在干净数据库完成全新安装测试；
3. 从上一插件版本完成升级测试；
4. 完成停用、重新启用和回滚演练；
5. 检查发行包中没有密钥和本地配置；
6. 生成 `manifest.json`、变更日志和摘要；
7. 创建插件版本标签；
8. 上传发行包和包摘要；
9. 发布安装步骤、兼容核心版本和已知限制。

插件独立仓库的标签建议使用：

```text
v1.0.0
```

如果多个插件共用一个仓库，标签建议使用：

```text
news-v1.0.0
shop-v1.0.0
```

不得只发布“最新版”而不提供固定版本、核心兼容版本和摘要。

## 11. 使用方安装插件

### 11.1 安装前检查

1. 确认接口端和管理端核心版本均为 `v0.0.2`；
2. 阅读插件 Manifest、README、变更日志和许可证；
3. 检查插件 ID 是否与已安装插件冲突；
4. 检查所有依赖插件及精确版本；
5. 验证发行包 SHA256；
6. 审查 Go、Vue、SQL 和依赖变化；
7. 确认没有密钥、远程代码加载或危险系统命令；
8. 备份数据库、接口端构建产物和管理端静态资源；
9. 先在测试环境安装。

### 11.2 合并插件源码

从使用方当前正在部署的代码创建安装分支。如果当前环境是没有其他改动的纯净 `v0.0.2`，可以直接以标签为基线：

```bash
git switch -c install/news-v1.0.0 v0.0.2
```

如果当前环境已经安装其他插件或包含补丁，应先切换到实际部署分支，再创建安装分支：

```bash
git switch production
git pull --ff-only
git switch -c install/news-v1.0.0
```

将发行包内容复制到对应目录：

```text
server_api/internal/plugins/news
server_api/sql/plugins/news
admin_client/src/plugins/news
```

人工合并 `register.patch`，确认 `news.New()` 已加入现有 `builtins`，同时保留其他插件注册项。

不要直接覆盖：

- `internal/plugins/register.go`；
- 核心 Router；
- 核心登录、权限和上传代码；
- 管理端全局样式；
- 已存在的其他插件目录。

### 11.3 执行数据库脚本

```bash
mysql -uroot -p production_database < server_api/sql/plugins/news/v1.0.0/migration.sql
mysql -uroot -p production_database < server_api/sql/plugins/news/v1.0.0/register.sql
```

执行后检查：

```sql
SELECT plugin_id, name, version, logo, author, homepage, status
FROM sys_plugin
WHERE plugin_id = 'news';

SELECT plugin_id, version, checksum, status, executed_at
FROM sys_plugin_migration
WHERE plugin_id = 'news'
ORDER BY executed_at;
```

新插件记录必须保持 `status = 2`，此时即使旧服务仍在运行，也不会尝试加载尚未发布的插件。

### 11.4 构建

接口端：

```bash
cd server_api
gofmt -l .
go vet ./...
go test ./...
go build -o bin/server_api .
```

管理端：

```bash
cd admin_client
pnpm install --frozen-lockfile
pnpm exec vue-tsc --noEmit
pnpm build
```

如果插件没有新增依赖，不应出现无关的 `go.sum` 或锁文件大范围变化。

### 11.5 发布并启用

推荐顺序：

1. 停止管理端 API 和用户端 API；
2. 再次确认数据库备份；
3. 发布包含插件的新后端二进制；
4. 发布包含插件页面的新管理端静态资源；
5. 将 `sys_plugin.status` 改为 `1`，或先启动管理端后在插件管理页面点击启用；
6. 启动管理端 API 和用户端 API；
7. 检查启动日志没有版本、迁移或依赖错误；
8. 重新登录管理端；
9. 给普通角色分配插件菜单及按钮权限；
10. 完成核心功能和插件功能冒烟测试。

如果服务启动失败，不要绕过注册中心校验。应根据错误检查未编译插件、版本、迁移摘要或依赖状态。

## 12. 插件升级

以 `news 1.0.0 → 1.1.0` 为例。

### 12.1 开发升级版本

从已经发布的插件版本分支或标签创建：

```bash
git switch -c plugin/news-v1.1.0 news-v1.0.0
```

如果插件 `1.1.0` 改为依赖新的核心能力，应从对应核心标签重新移植，并更新 `CoreVersion`。

新增：

```text
sql/plugins/news/v1.1.0/
├── migration.sql
└── register.sql
```

禁止修改 `v1.0.0/migration.sql`。

### 12.2 向前兼容顺序

推荐使用扩展—迁移—收缩策略：

1. 先新增允许旧代码继续工作的表、字段和索引；
2. 发布可同时兼容旧结构和新结构的代码；
3. 迁移或回填数据；
4. 确认所有实例升级完成；
5. 至少延后一个发布周期删除旧字段。

不要在同一次无停机升级中先删除旧代码仍在使用的字段。

### 12.3 使用方升级流程

1. 停用插件并重启服务；
2. 备份数据库和当前插件源码；
3. 审查新发行包及摘要；
4. 合并新后端和管理端源码；
5. 执行 `v1.1.0/migration.sql`；
6. 执行 `v1.1.0/register.sql`，写迁移记录并更新 `sys_plugin.version`；
7. 重新构建后端和管理端；
8. 发布新构建产物；
9. 在插件管理页面确认程序版本与安装版本一致；
10. 启用插件并重启两个 API 服务；
11. 完成升级验证。

## 13. 停用、回滚和卸载

### 13.1 停用

在插件管理页面点击停用，然后重启管理端 API 和用户端 API。

停用只停止插件路由和后台任务，不删除：

- 插件业务表；
- 插件上传文件；
- 菜单和角色授权；
- 插件迁移记录；
- 插件源码。

如果其他启用插件依赖当前插件，系统会拒绝停用。

### 13.2 代码回滚

回滚前确认数据库结构仍与旧代码兼容。推荐：

1. 停用插件；
2. 重启并确认核心服务正常；
3. 发布上一个插件版本的后端和前端构建；
4. 仅在明确支持时执行数据回滚 SQL；
5. 将 `sys_plugin.version` 恢复到与代码一致的状态；历史成功迁移记录保留，不伪造、不删除；
6. 再次启用并重启验证。

不要只回滚二进制而保留不兼容的插件版本记录，否则注册中心会拒绝启动。

### 13.3 卸载

卸载属于破坏性操作，默认不推荐。必须：

1. 停用插件并重启；
2. 确认没有其他插件依赖；
3. 导出插件业务数据和上传文件；
4. 从角色中移除插件权限；
5. 人工审查 `uninstall.sql`；
6. 明确批准后再删除菜单、业务表和数据；
7. 从 `RegisterBuiltins` 移除插件；
8. 删除后端和管理端插件源码；
9. 重新构建并发布；
10. 根据审计要求决定是否保留 `sys_plugin` 和迁移记录。

生产环境不应由 Web 页面自动执行 `DROP TABLE` 或删除插件文件。

## 14. CI/CD 建议

插件发行流水线至少包含：

1. 校验分支基线来自声明的核心 Tag；
2. 检查 Manifest 的插件 ID、版本和核心版本；
3. 检查插件表名前缀；
4. 检查 SQL 表和字段注释及 `utf8mb4_general_ci`；
5. 校验迁移文件 SHA256 与代码声明一致；
6. 执行后端格式化、测试、Vet 和构建；
7. 执行管理端类型检查和生产构建；
8. 扫描密钥、证书和高危依赖；
9. 在空数据库测试全新安装；
10. 从上一版本数据库测试升级；
11. 生成发行包和摘要；
12. 只有全部通过后才允许创建插件 Tag。

## 15. 开发者交付清单

- [ ] 接口端和管理端均基于同一核心 `v0.0.2` 标签。
- [ ] 分支名符合 `plugin/<plugin_id>-v<version>`。
- [ ] `plugin_id` 全局唯一且发布后不变。
- [ ] Manifest 包含版本、Logo、作者、主页、描述、核心版本和依赖。
- [ ] 后端、管理端和 SQL 目录符合规范。
- [ ] 已显式注册插件，但未覆盖其他插件注册项。
- [ ] 管理端页面使用 `/plugin/{plugin_id}/{view_path}`。
- [ ] migration.sql 不包含自身迁移摘要记录。
- [ ] SHA256 已同步到代码、register.sql 和 checksums.txt。
- [ ] 新安装默认状态为停用。
- [ ] 安装、升级、停用、回滚和卸载说明完整。
- [ ] 后端和管理端全部检查通过。
- [ ] 发行包不包含配置、密钥、构建缓存和无关代码。
- [ ] 已在空数据库和上一版本数据库完成测试。

## 16. 安装人员检查清单

- [ ] 已验证核心版本和插件兼容版本。
- [ ] 已验证发行包 SHA256。
- [ ] 已审查源码、SQL、依赖和许可证。
- [ ] 已确认插件 ID 和表名无冲突。
- [ ] 已备份数据库和当前构建产物。
- [ ] 已先在测试环境安装。
- [ ] 已人工合并 `register.patch`，没有覆盖其他插件。
- [ ] 已先保持插件停用并完成构建。
- [ ] 已检查迁移状态和摘要。
- [ ] 已发布后端和管理端构建产物。
- [ ] 已启用插件并重启两个 API 服务。
- [ ] 已分配角色权限并完成冒烟测试。
- [ ] 已准备明确的回滚方案。

## 17. 常见问题

### 为什么不能在管理页面直接上传并安装插件？

当前插件包含 Go 后端代码，必须经过源码审查和编译。允许 Web 页面写入并执行未知代码会显著扩大供应链和远程代码执行风险。

### 为什么启用或停用后必须重启？

Gin 路由和插件后台任务在进程启动阶段注册。`v0.0.2` 不做运行时路由卸载和热加载，以保持行为可预测。

### 为什么数据库有插件记录，页面却显示“未编译”？

插件 SQL 已安装，但插件源码没有加入当前后端构建，或者没有在 `RegisterBuiltins` 注册。必须完成源码合并和重新构建。

### 为什么服务提示安装版本和程序版本不一致？

`sys_plugin.version` 与 `Manifest.Version` 不同。检查是否漏执行升级 SQL、发布了错误二进制，或错误修改了版本记录。

### 为什么提示迁移摘要不一致？

已经登记的迁移文件被修改，或者执行的 SQL 与发行包不一致。不要直接修改数据库摘要绕过校验，应恢复官方迁移文件并重新核对安装过程。

### 插件可以只包含后端或只包含管理端吗？

可以。发行清单必须明确 `has_server` 和 `has_admin_client`。只有管理端页面而没有配套接口时仍需说明数据来源和权限边界；只有后端时管理端目录可以省略。

### 两个插件修改 register.go 冲突怎么办？

人工合并 import 和 `builtins` 列表，保留所有插件。禁止直接用某个插件发行包里的完整文件覆盖当前文件。

### 核心升级到新版本后插件怎么办？

从新的核心 Tag 创建兼容分支，完成移植、测试并发布新的插件版本，同时更新 `CoreVersion`。不要仅修改数据库中的核心兼容版本声明。
