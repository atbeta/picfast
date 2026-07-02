# Release 翻译参考（给 LLM，非用户文档）

维护者可在发版前更新本文件，帮助 Release 中译统一产品名、功能名与语气。

## 产品

- **PicFast**：自托管图床/图片托管服务；支持 S3/本地/云存储、CDN 加速、图片实时处理。
- 英文 slogan：**Self-hosted image hosting, fast and simple** → 可译为「自托管图床，快速简洁」

## 技术与栈（保留英文）

- **Go**、**Chi**、**pgx/v5**、**sqlc**、**JWT**、**PostgreSQL** — 一般保留英文。
- **React**、**TypeScript**、**Vite**、**Tailwind CSS** — 保留英文。
- **S3**、**MinIO**、**Cloudflare R2** — 保留英文。
- **MCP** / **Model Context Protocol** — 可写「MCP（模型上下文协议）」，或保留 MCP。
- **CSP** / **Content-Security-Policy** — 保留英文。
- **SSO** / **OAuth** / **OIDC** — 保留英文。
- **Webhook** — 可译为「Webhook / 网络钩子」，首次出现时标注一下。

## 功能模块（与 UI / 文档对齐）

- **Image Upload** — 图片上传
- **Album** — 相册
- **Group** — 用户组
- **Strategy** — 上传策略（指定存储位置、格式转换、水印等）
- **Pipeline** — 图片处理流水线
- **Audit Log** — 审计日志
- **API Token** — API 令牌
- **Webhook** — Webhook / 网络钩子
- **Personalization** — 个性化设置
- **Theme** — 主题
- **Dashboard** — 仪表盘
- **Console** — 控制台
- **Guest Upload** — 游客上传
- **Image Processing** — 图片处理（支持剪裁、缩放、水印、格式转换等）
- **X-Forwarded-For** — 保留英文，指反向代理传递的客户端 IP

## 常见动作 / 设置项

- **custom_css** — 自定义 CSS
- **site-level / user-level** — 站点级 / 用户级
- **light/dark/system mode** — 亮色/暗色/跟随系统
- **on-the-fly processing** — 实时处理 / 即时处理
- **link_mode=proxy** — 代理模式链接

## 语气

- 面向开发者与运维：简洁、专业；技术术语准确，不夸大营销。
- 安全、隐私相关表述需准确（如「JWT 认证」「API 令牌」「权限检查」等）。
- 不兼容变更必须标注「⚠️ 不兼容变更」，并提供升级指引。
