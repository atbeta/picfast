# PicFast Web

PicFast 的前端控制台与公共站点，基于 React 19、TypeScript、Vite、React Router 与 Tailwind CSS v4。

## 主要页面

- 公共上传页
- 登录 / 注册
- 用户控制台：上传、图片、相册、API Token、个人设置
- 管理后台：概览、用户、分组、存储策略、图片、系统设置

## 开发

```bash
pnpm install
pnpm dev
```

默认开发地址为 [http://localhost:5173](http://localhost:5173)。

Vite 会将以下请求代理到后端：

- `/api`
- `/i`
- `/t`

默认代理目标为 `http://localhost:8080`（与本地 `go run ./cmd/picfast` 一致），可通过环境变量 `VITE_BACKEND_URL` 覆盖（见 `vite.config.ts`）。

## 构建

```bash
pnpm build
```

默认产物目录为 `web/dist`。

根目录的 Go 服务在未显式配置 `PICFAST_SERVER_WEB_DIR` 时，也会尝试从 `web/dist` 加载静态资源。

## 质量检查

```bash
pnpm lint
pnpm exec tsc -b --pretty false
```
