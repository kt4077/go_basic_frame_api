# 插件开发规范

新开发者从核心版本建分支、制作发行包及安装升级的完整流程，参见 [插件开发、分发、安装与更新指南](./plugin-development-distribution-guide.md)。

## 1. 适用范围

本文档适用于基于当前项目开发的源码级业务插件，例如商城、新闻资讯、内容管理和营销活动等插件。

插件必须遵循现有项目的代码分层、数据库、接口安全和管理端 UI 规范。插件不是独立于主项目的另一套框架，不得复制认证、响应、上传、配置等基础设施。

当前插件核心契约版本为 `0.0.2`，采用编译期注册方式：

- 不支持上传 Go 文件后动态加载；
- 不支持运行中热启停；
- 插件状态、版本或依赖变更后必须重启管理端 API 和用户端 API；
- 插件后端和管理端页面都必须进入主项目构建流程。

## 2. 强制原则

1. 插件 ID 发布后不得修改，必须以小写字母开头，只能包含小写字母、数字和下划线，长度为 2～64 位。
2. 插件之间只能通过公开服务接口或明确声明的依赖协作，禁止直接修改其他插件的业务表。
3. Controller 只负责参数绑定、调用 Logic 和输出统一响应，业务规则及事务必须放在 Logic。
4. Logic 不得直接返回 Model，接口响应必须在 `resp` 中重新定义。
5. Model、Param、Resp 的字段必须添加 JSON、GORM 或校验 Tag，并写明字段注释。
6. 表关联关系在 Resp 中定义，不在 Model 中定义；适合关联加载的查询优先使用 GORM `Preload`。
7. 管理端业务接口默认注册到 `Permission` 路由组，禁止为了方便绕过登录、权限和操作日志。
8. 查询操作日志的接口必须关闭操作日志记录，避免查询行为递归产生新日志。
9. 数据库表和字段必须有注释，字符集统一为 `utf8mb4`，排序规则统一为 `utf8mb4_general_ci`。
10. 数据库枚举值必须从 `1` 开始，并在 Go 枚举文件和前端枚举文件中集中定义。
11. 普通业务表中的文件字段只保存相对路径，通过统一文件地址补全方法返回绝对地址；`sys_upload_file` 按现有规则同时保存相对路径和绝对路径。
12. SQL 不使用 GORM `AutoMigrate` 自动变更生产数据库，不在服务启动时执行 DDL。
13. 前端必须使用 TypeScript、ES6 语法和箭头函数，不得使用 `any` 绕过类型检查。
14. 插件页面必须复用现有组件、主题变量、间距、按钮、表格和弹窗规范，不得覆盖全局样式。
15. 时间展示精确到秒，不显示毫秒。

## 3. 标准目录结构

### 3.1 后端目录

```text
server_api/internal/plugins/{plugin_id}/
├── plugin.go
├── enums/
│   └── *.go
├── model/
│   └── *.go
├── service/
│   └── *.go
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

目录职责：

| 目录 | 职责 |
| --- | --- |
| `plugin.go` | 插件清单、迁移声明、路由注册和生命周期 |
| `enums` | 插件业务枚举，枚举值从 1 开始 |
| `model` | 插件自身的数据库表模型，不定义关联对象 |
| `service` | 管理端、用户端或其他插件共同使用的插件公开业务能力 |
| `admin` | 管理端接口，必须使用现有管理端鉴权体系 |
| `api` | 用户端接口，必须使用现有用户端鉴权体系 |

只有单端能力时可以省略对应的 `admin` 或 `api` 目录，但 `Plugin` 接口中的路由注册方法仍需实现并返回 `nil`。

Param 和 Resp 必须按业务功能拆分文件。例如：

```text
admin/param/article_create.go
admin/param/article_list.go
admin/resp/article_detail.go
admin/resp/article_list.go
```

禁止把整个插件的请求或响应结构全部堆放在一个 `param.go`、`resp.go` 中。

### 3.2 管理端目录

```text
admin_client/src/plugins/{plugin_id}/
├── api/
├── components/
├── enums/
├── types/
└── views/
    └── {view_path}/
        └── index.vue
```

数据库菜单路径：

```text
/plugin/{plugin_id}/{view_path}
```

页面文件：

```text
src/plugins/{plugin_id}/views/{view_path}/index.vue
```

例如 `/plugin/news/article` 对应 `src/plugins/news/views/article/index.vue`。

### 3.3 SQL 目录

每个插件使用独立版本目录：

```text
server_api/sql/plugins/{plugin_id}/
├── v1.0.0/
│   ├── migration.sql
│   ├── register.sql
│   └── uninstall.sql
└── v1.1.0/
    ├── migration.sql
    └── register.sql
```

`migration.sql` 只包含业务结构和数据迁移，`register.sql` 负责登记迁移摘要及插件信息，避免迁移文件包含自身摘要形成循环。`uninstall.sql` 只作为人工卸载参考，不允许应用程序自动执行。涉及删除表、字段或数据的语句必须带有醒目的危险操作说明。

## 4. 插件清单和契约

插件必须实现 `internal/common/plugin.Plugin`：

```go
package news

import (
    "context"

    commonplugin "server_api/internal/common/plugin"
)

type Plugin struct{}

func New() *Plugin {
    return &Plugin{}
}

func (p *Plugin) Manifest() commonplugin.Manifest {
    return commonplugin.Manifest{
        PluginID:    "news",
        Name:        "新闻资讯",
        Version:     "1.0.0",
        Logo:        "plugins/news/logo.png",
        Author:      "项目维护者",
        Homepage:    "https://example.com/plugins/news",
        Description: "提供新闻分类、文章和发布管理能力",
        CoreVersion: commonplugin.CoreVersion,
        Dependencies: []commonplugin.Dependency{},
    }
}

func (p *Plugin) Migrations() []commonplugin.Migration {
    return []commonplugin.Migration{
        {
            Version:     "1.0.0",
            Description: "创建新闻资讯基础表和菜单",
            Checksum:    "安装 SQL 文件的 SHA256 摘要",
        },
    }
}

func (p *Plugin) RegisterAdminRoutes(
    ctx *commonplugin.Context,
    groups commonplugin.AdminRouteGroups,
) error {
    return nil
}

func (p *Plugin) RegisterAPIRoutes(
    ctx *commonplugin.Context,
    groups commonplugin.APIRouteGroups,
) error {
    return nil
}

func (p *Plugin) Start(ctx context.Context, service commonplugin.ServiceType) error {
    return nil
}

func (p *Plugin) Stop(ctx context.Context) error {
    return nil
}
```

清单要求：

- `PluginID` 必须符合命名规则且全局唯一；
- `Name` 和 `Version` 不得为空；
- `Logo` 只保存文件相对路径，查询展示时由统一文件地址方法补全；
- `Author` 填写插件作者或维护团队，`Homepage` 填写可公开访问的插件主页；
- `Description` 应说明插件功能、适用场景和关键依赖；
- `Version` 推荐使用语义化版本，例如 `1.2.0`；
- `CoreVersion` 必须填写当前 `commonplugin.CoreVersion`；
- 所有依赖必须在 `Dependencies` 中声明；
- 当前依赖版本使用精确匹配，不支持 `>=1.0.0` 等范围表达式；
- 禁止形成循环依赖。

插件必须在 `internal/plugins/register.go` 显式注册：

```go
builtins := []commonplugin.Plugin{
    news.New(),
}
```

数据库启用但没有编译进程序的插件会导致服务拒绝启动，这是防止未知代码和不完整发布的安全保护。

## 5. 数据库规范

### 5.1 表和字段命名

插件业务表统一使用：

```text
plg_{plugin_id}_{table_name}
```

示例：

```text
plg_news_category
plg_news_article
plg_shop_product
plg_shop_order
```

禁止创建 `sys_` 前缀的插件业务表。`sys_` 只用于核心系统表。

### 5.2 建表要求

```sql
CREATE TABLE IF NOT EXISTS `plg_news_article` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `created_at` datetime(3) NOT NULL COMMENT '创建时间',
  `updated_at` datetime(3) NOT NULL COMMENT '更新时间',
  `deleted_at` datetime(3) DEFAULT NULL COMMENT '删除时间',
  `title` varchar(200) NOT NULL COMMENT '文章标题',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：1草稿，2已发布，3已下线',
  PRIMARY KEY (`id`),
  KEY `idx_news_article_deleted_at` (`deleted_at`),
  KEY `idx_news_article_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='新闻文章表';
```

必须满足：

- 表和每个字段都有明确注释；
- 主键、唯一约束、查询条件和排序字段根据访问方式建立索引；
- 业务唯一性必须由唯一索引兜底，不能只依赖代码先查询；
- 金额使用整数最小货币单位或明确精度的 `decimal`，禁止使用浮点数；
- 并发更新使用事务、原子条件更新或锁，禁止“先查再无条件写”；
- 外键约束是否使用应与现有项目保持一致，但关联 ID 必须有索引；
- SQL 应具备可重复执行能力，或在文档中明确只能执行一次；
- 禁止在升级脚本中静默删除业务数据。

### 5.3 Model 规范

```go
// Article 新闻文章模型。
type Article struct {
    ID        uint64         `gorm:"column:id;primaryKey" json:"id" comment:"文章ID"`
    CreatedAt time.Time      `gorm:"column:created_at" json:"created_at" comment:"创建时间"`
    UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at" comment:"更新时间"`
    DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-" comment:"删除时间"`
    Title     string         `gorm:"column:title;size:200;not null" json:"title" comment:"文章标题"`
    Status    uint8          `gorm:"column:status;not null" json:"status" comment:"文章状态"`
}

// TableName 返回新闻文章表名。
func (Article) TableName() string {
    return "plg_news_article"
}
```

Model 只映射当前表字段，不定义分类、作者等关联对象。关联响应在 Resp 中定义。

### 5.4 迁移记录

1. 执行插件 SQL。
2. 计算 SQL 文件 SHA256 摘要。
3. 向 `sys_plugin_migration` 写入状态为 `1` 的成功记录。
4. `Migrations()` 中填写相同版本和摘要。
5. 安装或升级全部完成后，再更新 `sys_plugin.version`。

迁移失败时必须记录状态 `2` 和错误原因，不得把未完成的迁移标记为成功。修改已经发布的 SQL 会造成摘要不一致；已发布迁移不得修改，只能新增更高版本迁移。

## 6. 后端分层规范

### 6.1 Param

```go
// ArticleCreate 创建文章参数。
type ArticleCreate struct {
    CategoryID uint64 `json:"category_id" binding:"required" comment:"分类ID"`
    Title      string `json:"title" binding:"required,max=200" comment:"文章标题"`
    Content    string `json:"content" binding:"required" comment:"文章内容"`
}
```

- 必须定义 JSON、校验和 comment Tag；
- 参数校验在进入 Logic 前完成；
- ID、状态、分页大小等必须校验边界；
- 不允许客户端提交创建人、租户、权限范围等应由服务端确定的字段。

### 6.2 Resp

```go
// ArticleDetail 文章详情响应。
type ArticleDetail struct {
    ID        uint64       `json:"id" gorm:"column:id" comment:"文章ID"`
    Title     string       `json:"title" gorm:"column:title" comment:"文章标题"`
    CoverPath string       `json:"cover_path" gorm:"column:cover_path" comment:"封面相对路径"`
    CoverURL  string       `json:"cover_url" gorm:"-" comment:"封面完整地址"`
    Category  CategoryItem `json:"category" gorm:"foreignKey:CategoryID" comment:"所属分类"`
    CreatedAt string       `json:"created_at" gorm:"column:created_at" comment:"创建时间"`
}
```

- Logic 返回 Resp，不直接暴露 Model；
- 关联对象和 GORM 预加载关系只定义在 Resp；
- 文件相对路径和完整访问地址应使用不同字段表达；
- 对外时间统一格式化到秒；
- 不返回内部错误、密码、密钥、完整存储凭据等敏感字段。

### 6.3 Logic 和事务

- Logic 负责权限范围、业务状态、事务、并发控制和 Resp 组装；
- 多表写入必须使用数据库事务；
- 事务内部所有查询和写入必须使用同一个事务对象；
- 外部网络调用不要长时间占用数据库事务；
- 重试操作必须设计幂等键或唯一约束；
- 列表查询必须分页并限制最大分页大小；
- 避免循环查询，关联数据优先批量查询或 `Preload`；
- 用户可控排序字段必须使用白名单，禁止直接拼接 SQL；
- 数据库错误通过项目统一错误封装处理，不把底层错误直接返回客户端。

### 6.4 Controller

- 只完成参数绑定、上下文信息获取、Logic 调用和统一响应；
- 不直接操作数据库；
- 不在 Controller 中编写复杂业务判断；
- 所有接口沿用项目现有响应结构和错误处理方式；
- 上传、文件地址补全、分页等能力复用公共封装。

## 7. 路由、鉴权和审计

### 7.1 管理端路由

```go
func (p *Plugin) RegisterAdminRoutes(
    ctx *commonplugin.Context,
    groups commonplugin.AdminRouteGroups,
) error {
    articleController := admincontroller.NewArticle(ctx.App)

    groups.Permission.GET("/plugin/news/article/list", articleController.List)
    groups.Permission.POST("/plugin/news/article/create", articleController.Create)
    groups.Permission.PUT("/plugin/news/article/update", articleController.Update)
    groups.Permission.DELETE("/plugin/news/article/delete", articleController.Delete)
    return nil
}
```

路由选择：

| 路由组 | 使用场景 | 要求 |
| --- | --- | --- |
| `Permission` | 管理端普通业务接口 | 默认选择，具备登录、接口权限和操作日志链路 |
| `Auth` | 已登录即可访问的个人能力 | 必须说明不需要角色权限的原因 |
| `Public` | 无法使用登录态的回调或公开能力 | 必须自行实现验签、防重放、限流和幂等 |

公开支付、短信、微信等回调必须验证第三方签名，并使用请求时间、随机数或事件 ID 防止重放。回调重复到达不能产生重复订单、重复退款或重复发放权益。

### 7.2 用户端路由

- 用户私有数据接口注册到 `Auth`；
- 只有确实公开的列表、详情或第三方回调才放入 `Public`；
- 即使已登录，也必须在 Logic 中校验资源归属；
- 不得只根据客户端传入的用户 ID 查询或修改数据。

### 7.3 菜单和权限

- 菜单、按钮和接口权限 SQL 与插件安装 SQL 一起交付；
- 页面路径使用 `/plugin/{plugin_id}/{view_path}`；
- API 权限标识必须与实际请求方法和路径一致；
- 普通角色默认不自动获得新插件权限；
- 超级管理员和普通管理员的行为沿用现有权限体系；
- 菜单 ID、父子关系和唯一索引必须在目标数据库中确认，禁止复制固定 ID 后直接覆盖现有菜单；
- 权限变更后按现有机制清理缓存或重新登录验证。

## 8. 前端开发规范

1. 使用 Vue 3、`<script setup lang="ts">`、ES6 和箭头函数。
2. 网络请求统一使用现有 `request<T>()`，接口域名只从环境变量和统一请求封装读取。
3. 插件 API、类型、枚举、组件和页面按目录拆分，禁止在页面中混写大量请求与数据转换逻辑。
4. 页面按钮必须使用现有 `v-perm` 权限指令。
5. 表格、表单、分页、上传、空状态和错误提示复用现有组件及交互规则。
6. 页面不得定义影响其他模块的全局 CSS；样式优先使用 `scoped` 和现有主题变量。
7. 不硬编码后端域名、存储域名、权限值或中文状态映射。
8. 文件上传只向业务接口提交相对路径；展示地址使用后端补全结果。
9. 日期时间统一展示到秒。
10. 新增页面后执行类型检查和生产构建，确认动态路由可以解析。

## 9. 生命周期和后台任务

`Start` 只用于启动确实需要的后台任务、订阅或连接。普通 CRUD 插件应直接返回 `nil`。

后台任务必须：

- 监听传入的 `context.Context`；
- 在 Context 取消后及时退出；
- 避免重复启动；
- 对共享状态加锁或使用安全的并发结构；
- 设置外部请求超时；
- 对可重试任务设置退避和最大重试次数；
- 对消费任务设计幂等处理；
- 在 `Stop` 中释放插件独占资源；
- 不关闭宿主提供的数据库或 Redis 连接。

插件按依赖顺序启动，按相反顺序停止。任一插件启动失败时，服务启动失败，已启动插件会按相反顺序停止。

## 10. 配置和敏感信息

- 插件配置应使用独立配置键或独立业务表，避免无限扩展核心配置结构；
- 密钥、证书、支付私钥、短信凭证不得写入源码、日志或前端；
- 敏感配置应加密存储或通过安全环境变量注入；
- 返回配置详情时必须脱敏，更新接口使用“留空表示不修改”等明确语义；
- 配置变更需要考虑缓存失效和并发读取一致性；
- 多实例部署时不得依赖单机内存保存关键状态。

## 11. 依赖和插件间协作

- 插件不得导入其他插件的 Controller、Param、Resp 或 Model；
- 可复用能力应由被依赖插件在 `service` 中提供稳定接口；
- 必须在 `Manifest.Dependencies` 中声明依赖插件和精确版本；
- 禁止跨插件直接写表；确需读取时也应优先通过服务接口；
- 通用且不包含业务语义的能力放入 `pkg`；
- 多业务模块共同依赖的业务能力放在合适的 `internal/common` 子目录；
- 不得为了单个插件把插件专属逻辑放入 `common`。

## 12. 安装、升级和卸载流程

### 12.1 安装

1. 审查插件源码和依赖。
2. 备份数据库及当前构建产物。
3. 依次执行插件 `migration.sql` 和 `register.sql`。
4. 写入成功迁移记录和停用状态的 `sys_plugin` 记录。
5. 将插件后端加入 `RegisterBuiltins`。
6. 将管理端插件页面加入源码并完成构建。
7. 在测试环境启动并验证。
8. 更新 `sys_plugin.status = 1`。
9. 重启管理端 API 和用户端 API。
10. 分配菜单权限并完成冒烟验证。

### 12.2 升级

1. 禁止修改历史迁移文件。
2. 新增版本目录和 `upgrade.sql`。
3. 先执行数据库向前兼容变更，再发布可兼容新旧结构的代码。
4. 写入迁移成功记录并更新插件版本。
5. 同时发布后端和管理端构建产物。
6. 重启服务并检查版本、依赖和迁移校验。
7. 涉及删除旧字段时，至少延后一个发布周期处理。

### 12.3 停用和卸载

- 停用只修改插件状态并重启服务，不删除业务数据；
- 卸载前确认没有其他插件依赖；
- 默认保留插件表、上传文件和迁移记录；
- 删除数据、菜单、文件和表必须由管理员明确执行并提前备份；
- 回滚代码前确认数据库结构和插件版本与旧代码兼容。

## 13. 发布前检查清单

### 后端

- [ ] 插件 ID、版本、核心版本和依赖声明正确。
- [ ] 已在 `internal/plugins/register.go` 注册。
- [ ] Controller、Logic、Param、Resp、Model 分层符合规范。
- [ ] Logic 没有直接返回 Model。
- [ ] 关联关系只在 Resp 定义，并检查了 N+1 查询。
- [ ] 管理端业务接口默认使用 `Permission` 路由组。
- [ ] Public 接口已完成验签、防重放、限流和幂等。
- [ ] 数据范围和资源归属在服务端校验。
- [ ] 并发写入有唯一索引、条件更新、事务或锁保护。
- [ ] 后台任务支持 Context 取消和安全停止。
- [ ] 日志中没有密码、Token、密钥和完整敏感配置。

### 数据库

- [ ] 表名使用 `plg_{plugin_id}_` 前缀。
- [ ] 表和字段都有注释。
- [ ] 字符集为 `utf8mb4`，排序规则为 `utf8mb4_general_ci`。
- [ ] 枚举从 1 开始，并已同步后端及前端枚举。
- [ ] 唯一性、关联字段、筛选和排序字段索引合理。
- [ ] SQL 不会删除或覆盖现有业务数据。
- [ ] 迁移版本和 SHA256 摘要已写入迁移记录。
- [ ] 菜单和权限 SQL 不与现有数据冲突。

### 管理端

- [ ] 页面路径和文件目录满足插件页面解析约定。
- [ ] 使用 TypeScript、ES6 和箭头函数。
- [ ] 请求通过统一封装，未硬编码接口域名。
- [ ] 操作按钮使用 `v-perm`。
- [ ] UI、明暗主题和响应式布局符合现有页面风格。
- [ ] 时间显示到秒，文件地址使用后端补全结果。
- [ ] 页面无跨模块全局样式污染。

### 验证命令

后端：

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

所有检查通过后，还必须在测试环境验证插件安装、启用、权限分配、停用、升级和服务重启流程。

## 14. 禁止事项

- 禁止插件绕过注册中心直接修改核心路由初始化。
- 禁止在插件启动时自动建表或修改表结构。
- 禁止把 Model 直接作为接口响应。
- 禁止在 Model 中定义业务关联结构。
- 禁止跨插件直接修改数据表。
- 禁止使用未校验的字符串拼接 SQL、排序或表名。
- 禁止在 Public 路由中信任客户端身份字段。
- 禁止启动不能停止、没有超时或无限重试的 Goroutine。
- 禁止把第三方密钥、Token、证书或用户敏感信息写入日志。
- 禁止插件覆盖全局 CSS、全局组件或核心菜单行为。
- 禁止修改已经发布并登记摘要的迁移 SQL。
- 禁止未经备份直接执行插件卸载或破坏性回滚。
