# 请求参数验证

项目使用 `pkg/validate` 统一完成请求绑定、规则校验和错误文案转换。管理端、用户端和插件 Controller 不再直接调用 Gin 的 `ShouldBindJSON`、`ShouldBindQuery` 或 `ShouldBindUri`。

## Param 标签

每个请求字段都必须同时声明字段映射、语义标签和注释：

```go
type AccountUpdateReq struct {
    Account string `json:"account" binding:"required,min=4,max=32" validate:"登录账号" comment:"登录账号"`
}
```

| 标签 | 用途 |
| --- | --- |
| `json` / `form` / `uri` | 请求字段映射 |
| `binding` | 校验规则，沿用 Gin/go-playground validator 规则 |
| `validate` | 面向用户的字段语义，只写简洁名称，不写校验规则或完整句子 |
| `comment` | 完整字段说明，可包含枚举、限制和补充语义 |

`validate` 通常从 `comment` 的核心名称提取。例如 `comment:"目标状态：1启用，2禁用"` 对应 `validate:"目标状态"`。所有字段都需要声明 `validate`，即使当前没有 `binding` 规则，也便于后续增加规则时获得稳定错误文案。

## Controller 调用

JSON 请求体和查询参数统一调用 `Bind`，工具会根据请求方法选择现有绑定方式：

```go
var req param.MemberSetStatusReq
if err := requestvalidate.Bind(c, &req); err != nil {
    response.Fail(c, response.CodeErrParams, err.Error())
    return
}
```

URI 参数使用 `BindURI`：

```go
if err := requestvalidate.BindURI(c, &req); err != nil {
    response.Fail(c, response.CodeErrParams, err.Error())
    return
}
```

对已经完成赋值的结构体可以调用 `validate.Struct(&req)`。禁止在 Controller 重新解析 validator 原始错误，也不要返回统一的“参数错误”覆盖字段语义。

## 默认错误文案

验证器默认转换常用规则：

| binding 规则 | 示例返回信息 |
| --- | --- |
| `required` | `登录账号不能为空` |
| `oneof=1 2` | `账号状态取值不正确` |
| `len=6` | `短信验证码长度必须为6` |
| `min=6`（字符串） | `新密码长度不能少于6个字符` |
| `max=32`（字符串） | `昵称长度不能超过32个字符` |
| `min/max`（数组） | `日志ID列表至少包含1项` / `最多包含500项` |
| `email` | `邮箱格式不正确` |
| `url` | `插件主页地址必须是有效的URL地址` |
| `alphanum` | `分类标识只能包含字母和数字` |
| `eq=...` | `全量清空确认标识必须为...` |

JSON 语法错误统一返回“请求参数格式错误”；JSON 字段类型错误会尽量根据 `validate` 标签返回“字段语义格式不正确”。系统只返回第一条验证错误，避免向调用方一次暴露过多内部规则。

## 自定义规则

字段级规则通过 `RegisterValidation` 注册，结构体跨字段规则通过 `RegisterStructValidation` 注册。注册操作必须在服务启动阶段、处理并发请求之前完成：

```go
err := validate.RegisterValidation("slug", func(field playground.FieldLevel) bool {
    return slugPattern.MatchString(field.Field().String())
})
```

Param 中仍通过 `binding` 使用规则，并继续用 `validate` 定义字段语义：

```go
Slug string `json:"slug" binding:"required,slug" validate:"文章标识" comment:"文章标识"`
```

未内置专用文案的自定义规则会返回“文章标识不符合要求”或带规则参数的提示。业务唯一性、数据库状态、当前用户权限等需要访问外部状态的判断仍放在 Logic 层，不应注册为字段验证规则。
