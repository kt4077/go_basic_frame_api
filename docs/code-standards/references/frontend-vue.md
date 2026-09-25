# 前端 Vue / TypeScript 落地参考（admin_client）

## 目录结构

```
src/
├── api/         按业务模块封装请求（auth / user / member / role / menu / dept / sms / wechat / payment / storage / platform / upload / operation_log / dashboard）
├── types/       与 api 同名的请求与响应类型
├── enums/       前端枚举（common / menu / member / channel / storage / platform）
├── utils/       auth（token）、datetime、echarts、views
├── store/       Pinia：app / user / platform / tags
├── router/      静态路由 + 后端菜单驱动的动态路由
├── directives/  v-perm 等
├── components/  通用组件（如 AvatarUpload）
├── styles/      主题变量与公共样式
└── views/       config / dashboard / error / login / maintain / profile / redirect / system / user
```

页面路径与后端菜单 `path` 对应：`/system/user` → `src/views/system/user/index.vue`。

## 接口层

统一 `request<T>()`，禁止在页面直接使用 axios：

```ts
import { request } from './http'
import type { UserItem, UserListReq, UserSaveReq } from '@/types/user'
import type { PageResult } from '@/types/common'

export const getUserList = (params?: Partial<UserListReq>) => {
  return request<PageResult<UserItem>>({ url: '/admin/user/list', method: 'get', params })
}

export const createUser = (data: UserSaveReq) => {
  return request<UserItem>({ url: '/admin/user/add', method: 'post', data })
}
```

命名：`getXxxList` / `getXxxTree` / `createXxx` / `updateXxx` / `deleteXxx` / `saveXxx`。
无返回数据的操作返回 `request<null>`。

## 类型

公共类型（`src/types/common.ts`）优先复用：

```ts
ApiResponse<T>   // { code, msg, data }
PageQuery        // { page, page_size }
PageResult<T>    // { list, total }
TreeNode<T>      // { data: T, children?: TreeNode<T>[] }
BaseEntity       // { id, created_at, updated_at }
```

- 禁止 `any`，不确定用 `unknown`。
- 列表查询参数继承 `PageQuery`；列表响应用 `PageResult<T>`；实体继承 `BaseEntity`。
- 后端树为 `{ data, children }`，组件需要平铺时写转换函数并注释原因。
- 富文本编辑统一复用 `src/components/RichTextEditor.vue`，业务页面和插件只绑定内容与提示配置，不重复维护 WangEditor 上传、主题或销毁逻辑。

## 枚举

```ts
export const Status = { Enabled: 1, Disabled: 2 } as const
export const StatusLabels: Record<number, string> = { 1: '启用', 2: '禁用' }
```

数值必须与后端 `internal/common/enums` 一致，从 1 开始。

## 页面模板要点

```vue
<script setup lang="ts">
// 人员管理：人员 CRUD、角色分配、重置密码、踢下线
import { onMounted, reactive, ref } from 'vue'
import { getUserList } from '@/api/user'
import { formatDateTimeCell } from '@/utils/datetime'
import type { UserItem } from '@/types/user'
import type { PageResult } from '@/types/common'

const loading = ref(false)
const list = ref<UserItem[]>([])
const query = reactive({ keyword: '', page: 1, page_size: 20 })

const load = async () => {
  loading.value = true
  try {
    const res = await getUserList({ keyword: query.keyword || undefined, page: query.page, page_size: query.page_size })
    list.value = res.list
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>
```

- 使用 `<script setup lang="ts">`，组合式 API，函数统一箭头函数。
- 加载态在 `finally` 关闭，错误由 `http.ts` 拦截器统一提示。
- 公共样式：`.page-card`、`.toolbar`、`.table-operations`；主题变量取自 `src/styles/index.css`。
- 明暗主题都要可用，禁止硬编码仅适配白底的颜色。
- 日期时间用 `formatDateTime` / `formatDateTimeCell` 格式化到秒。
- 按钮权限：`v-perm="'POST:/admin/user/add'"`。
- 上传：提交业务数据时传 `relative_path`。

## 请求与环境

- `baseURL` 来自 `import.meta.env.VITE_ADMIN_API_BASE_URL`（`.env.development` / `.env.production`）。
- 拦截器自动携带 `Authorization: Bearer <token>`，业务码 401 或 HTTP 401 时清 token 并跳转登录页。
- 相同错误信息 2 秒内只提示一次。
- 环境变量必须以 `VITE_` 开头；禁止存放密钥。

## 提交前

```bash
pnpm exec vue-tsc --noEmit
pnpm build
```

确认：亮/暗主题正常、无权限角色看不到按钮且后端返回 403、空态与错误态处理正确、字段与后端一致、未提交 `.env.local` / 构建产物 / 密钥。
