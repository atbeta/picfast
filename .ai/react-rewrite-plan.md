# PicFast 前端 React 重写计划（单线程执行版）

## 当前状态（已完成）
- 已将旧前端备份为 `web-vue`。
- 已在 `web` 初始化 `React + Vite + TypeScript + Tailwind`。
- 已接入基础依赖：`react-router-dom`、`axios`、`@tanstack/react-query`、`i18next`、`react-i18next`、`react-hook-form`、`zod`。
- 已搭建基础骨架：
  - 路由分层：公开页 + `/console/*`
  - 主题切换：`light` / `dark` / `system`
  - 双语初始化：`zh-CN` / `en-US`
  - API 拦截器（含 token refresh）
- `pnpm build` 已通过。

## 目标与边界
- 保持后端 API 契约不变（`/api/v1`、`/i`、`/t`）。
- 前端定位：工具化图床（无图片流首页），`/` 为游客上传页。
- 主题只支持：`light` / `dark` / `system`。
- 语言支持：中文 + 英文。
- 最终在新前端稳定后删除 `web-vue`。

## 单线程执行步骤

### Step 1：认证闭环（优先）
- 实现登录/注册真实表单与接口联调。
- 完成登录态落地（token、refresh_token、登出）。
- 路由守卫完善（游客/登录用户/管理员）。
- 验收：
  - 可注册、登录、刷新后保持登录态；
  - token 失效后可 refresh 或回到登录页。

### Step 2：游客上传页可用化
- 实现上传组件（拖拽、点击上传、进度、成功结果）。
- 接通 `POST /api/v1/upload`。
- 展示上传结果（URL、Markdown、HTML 一键复制）。
- 验收：
  - 游客可上传；
  - 错误态（格式/大小/网络）有可读提示。

### Step 3：控制台核心页
- `/console/upload`：登录用户上传。
- `/console/images`：图片列表、基础操作（查看/复制/删除）。
- `/console/albums`：相册列表与基础增删改。
- `/console/api-tokens`：Token 创建、列表、删除。
- 验收：
  - 以上页面可正常请求 API；
  - 核心操作路径可走通。

### Step 4：设置与体验收口
- `/console/settings`：语言、主题、个人信息入口。
- 中英文案补齐（先核心页面，后长尾）。
- 统一空态、加载态、错误态与提示反馈。
- 验收：
  - 语言切换覆盖核心流程；
  - 主题切换与 system 跟随行为稳定。

### Step 5：管理员模块迁移
- `/console/admin/users`
- `/console/admin/groups`
- `/console/admin/strategies`
- `/console/admin/images`
- `/console/admin/settings`
- 验收：
  - 管理员路由可访问，普通用户不可访问；
  - 管理核心操作可用。

### Step 6：稳定化与清理
- 全面回归：认证、上传、列表、管理、双语、主题。
- 运行构建与基础检查，修复阻断问题。
- 确认无回退需求后删除 `web-vue`。

## 约定执行方式
- 一次只做一个 Step，完成后先验收再推进下一步。
- 每个 Step 至少包含：
  - 代码变更说明
  - 本地验证命令和结果
  - 风险/后续待办

## 下一步（立即执行）
- 进入 `Step 1：认证闭环`，先完成登录/注册页面与接口联调。
