# 后端 Go 落地参考（server_api）

## 目录结构与归属

```
cmd/        Cobra 子命令：service admin / service api / version（-c 指定配置）
config/     配置结构体、加载与默认值
router/     admin.go / api.go，两端路由分文件
internal/
  common/
    app/          配置 + GORM + Redis 生命周期、健康检查、必需表检查
    auth/         登录流水、JWT 签发解析、踢人下线
    enums/        业务枚举（按功能分文件）
    middleware/   两端共用：CORS、Auth
    model/        GORM 模型 sys_*
    upload/       上传业务、默认存储加载、本地文件访问
  admin/
    controller/ logic/ param/ resp/ middleware/ permission/
  api/
    controller/ logic/ param/ resp/
pkg/        response pagination password tree dberror oss
sql/        建表与升级脚本
docs/       接口文档与规范
```

归属判断顺序：两端共用 → `internal/common/<功能>`；仅一端使用 → `internal/<端>/<功能>`；无业务语义 → `pkg/<功能>`。

## 分层职责

- controller：只做 `ShouldBind*` → 调 logic → `response.OK/Fail`。
- logic：业务规则、事务、缓存维护；方法签名统一 `func (l *XxxLogic) Action(c *gin.Context, req *param.XxxReq) (*resp.XxxRes, error)`。
- param / resp：结构体定义，**按功能拆分文件**（`user.go`、`menu.go`…），禁止 `param.go` 聚合。
- model：仅放 GORM 模型与表名，不放响应结构。

## 代码模板

### param

```go
package param

type UserSaveReq struct {
    ID       uint   `json:"id" comment:"主键ID"`
    Username string `json:"username" comment:"登录账号"`
    Status   int    `json:"status" binding:"omitempty,oneof=1 2" comment:"状态"`
}
```

查询类（GET）字段补 `form:"..."`：

```go
type UserListReq struct {
    Keyword  string `form:"keyword" comment:"搜索关键字"`
    Page     int    `form:"page" comment:"页码"`
    PageSize int    `form:"page_size" comment:"每页数量"`
}
```

### resp

```go
package resp

type UserListRes struct {
    List  []UserItem `json:"list" comment:"用户列表"`
    Total int64      `json:"total" comment:"数据总数"`
}
```

实体统一内嵌 `BaseItem`（`id` / `created_at` / `updated_at`），树形接口由 `pkg/tree.BuildTree` 生成 `data` + `children`。

### logic

```go
// Create 新增用户：写主表、维护角色关联并清理权限缓存。
func (l *UserLogic) Create(c *gin.Context, req *param.UserSaveReq) (*resp.UserItem, error) {
    // 1. 校验业务规则
    // 2. 事务内写库
    // 3. 清理权限缓存
    // 4. 返回 resp.Item
}
```

### controller

```go
func (h *UserController) Create(c *gin.Context) {
    var req param.UserSaveReq
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Fail(c, response.CodeErrParams, "参数错误")
        return
    }
    res, err := h.Logic.Create(c, &req)
    if err != nil {
        response.Fail(c, response.CodeErrBusiness, err.Error())
        return
    }
    response.OK(c, res)
}
```

### 路由

```go
perm.POST("/user/add", userC.Create)   // 登录 + 接口级权限 + 操作日志
auth.POST("/logout", authC.Logout)     // 仅登录
pub.POST("/login", authC.Login)        // 公开
```

## 强制约定

1. controller 必须向 logic 传 `*gin.Context`；claims / IP / UA 在 logic 内获取。
2. 统一 `response.OK/Fail`；参数错误用 `CodeErrParams`，业务失败用 `CodeErrBusiness`。
3. 数据库错误用 `pkg/dberror` 识别（重复键 → 友好提示）。
4. 禁止 `AutoMigrate`，表变更提供 `sql/` 脚本。
5. 并发写入优先 `clause.OnConflict` upsert，并在数据库层加唯一约束。
6. 权限相关变更（菜单、角色、用户角色）后清理权限缓存。
7. 密钥类字段响应不返回；写入留空表示不修改。
8. 上传只存相对路径，地址动态解析；远程抓取复用 `safeRemoteHTTPClient`。
9. 数值枚举从 1 开始，`0` 表示未设置/不过滤。
10. 新增配置同步 `config/config.go`、`config.example.yaml`、README；新增接口同步 `docs/admin_openapi.yaml`。

## 提交前

```bash
gofmt -l .
go vet ./...
go build ./...
go test ./...
```
