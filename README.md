# PicFast

[![CI](https://github.com/atbeta/picfast/actions/workflows/ci.yml/badge.svg)](https://github.com/atbeta/picfast/actions/workflows/ci.yml)
[![Release Please](https://github.com/atbeta/picfast/actions/workflows/release-please.yml/badge.svg)](https://github.com/atbeta/picfast/actions/workflows/release-please.yml)
[![Publish Docker Image](https://github.com/atbeta/picfast/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/atbeta/picfast/actions/workflows/docker-publish.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/xbeta/picfast?logo=docker)](https://hub.docker.com/r/xbeta/picfast)
[![Docker Image](https://img.shields.io/docker/image-size/xbeta/picfast/latest?logo=docker&label=image)](https://hub.docker.com/r/xbeta/picfast)

[English](README.en.md)

PicFast 是一个面向个人与团队的现代化图床与图片托管服务，支持游客上传、多用户权限、多存储策略、管理后台以及 AI / MCP 集成能力。

## 界面预览 / UI Preview

| 中文界面 | English UI |
|---|---|
| ![PicFast Console Chinese UI](.github/assets/console-zh-cn.png) | ![PicFast Console English UI](.github/assets/console-en-us.png) |

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.26, Chi Router, pgx/v5, sqlc, JWT |
| 前端 | React 19, TypeScript, Vite, React Router, Tailwind CSS v4 |
| 数据库 | PostgreSQL 16 |
| 存储 | 本地文件系统 / S3 兼容对象存储 / 七牛云 Kodo / 阿里云 OSS / 腾讯云 COS / WebDAV |
| 可观测性 | Prometheus metrics, health check |

## 当前能力

- 用户注册、登录、刷新令牌、登出
- 游客上传与认证用户上传
- 图片列表、权限更新、删除
- 相册管理
- 管理后台：用户、分组、存储策略、系统设置、图片管理
- 本地、S3 兼容、七牛云 Kodo、阿里云 OSS、腾讯云 COS 与 WebDAV 存储策略
- 图片压缩、水印、同步缩略图生成
- 审核能力与审核状态回传
- ShareX 配置下载与上传接口
- API Token 与 MCP Server 集成
- 健康检查、Prometheus 指标（内部端口）、pprof 调试端点

## 存储策略

管理后台可创建并绑定多种存储策略。`strategy_type` 支持：

- `local`：本地文件系统，配置 `root` 与 `url`
- `s3`：S3 兼容对象存储，配置 `endpoint`、`region`、`bucket`、`access_key`、`secret_key`、`url`
- `kodo`：七牛云 Kodo，配置 `access_key`、`secret_key`、`bucket`、`domain`、`zone`、`private`
- `oss`：阿里云 OSS，配置 `endpoint`、`bucket`、`access_key`、`secret_key`、`url`
- `cos`：腾讯云 COS，配置 `bucket_url`、`secret_id`、`secret_key`、`url`
- `webdav`：WebDAV，配置 `endpoint`、`username`、`password`、`url`

其中 `url` / `domain` 用于生成公开访问地址；未配置公开地址时，部分策略会回退到服务端或存储端点地址。

## 快速开始

### 前置要求

- Go 1.26+
- Node.js 20+ 与 pnpm
- PostgreSQL 16

说明：应用启动时会自动执行数据库迁移；只有在需要手动调试迁移版本时才需要额外安装 `golang-migrate` CLI。

### 1. 启动数据库

```bash
make docker-up
```

或使用本地 PostgreSQL：

```bash
createdb picfast
```

### 2. 配置环境

```bash
cp .env.example .env
```

PicFast 同时支持两种配置方式：

- 环境变量：直接加载 `.env` 中的 `PICFAST_*` 配置
- YAML 文件：使用根目录 `config.example.yaml` 复制为 `config.yaml`

如果你更习惯文件配置，可以这样开始：

```bash
cp config.example.yaml config.yaml
```

环境变量优先级高于 `config.yaml`。通常本地开发二选一即可，不需要两边同时维护。

### 站点个性化、SEO 与统计

管理员后台的“站点设置”支持配置站点简介、页脚附加文案与链接、访问统计等。这些配置会持久化到数据库，并通过 `/api/v1/config` 下发给前端用于：

- 更新 `<title>`、`meta description`、Open Graph 和 Twitter 分享信息
- 在页脚按需显示两行自定义文本（可带可选 `http`/`https` 链接），以及 PicFast 版本和 GitHub 链接
- 按需注入 Plausible、Umami、Google Analytics 4、百度统计或自定义统计脚本

也可以通过配置文件或环境变量提供初始值：

```yaml
app:
  name: "PicFast"
  site_description: "Modern self-hosted image hosting for teams."
  footer_text_1: "京ICP备12345678号-1"
  footer_link_1: "https://beian.miit.gov.cn/"
  footer_text_2: "京公网安备11000002000001号"
  footer_link_2: "https://www.beian.gov.cn/"
  analytics_provider: "umami" # plausible | umami | ga4 | baidu | custom
  analytics_config: '{"script_url":"https://analytics.example.com/script.js","website_id":"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"}'
```

某行 `footer_text_*` 为空时，该行不会在页脚展示；`footer_link_*` 留空则该行仅显示纯文本。未配置统计服务时不会加载第三方脚本；页脚会默认显示 `Powered by PicFast`，GitHub 链接固定指向 `https://github.com/atbeta/picfast`。生产环境如果启用了严格 CSP，需要为所选统计服务补充相应的 `script-src` 白名单。

### 3. 可选：填充开发数据

```bash
make seed
```

数据库迁移会在后端启动时自动执行，不需要手动运行 `make migrate-up`。如果你只想使用空数据库体验首次初始化流程，可以跳过 seed。

默认会创建：

- 普通测试用户：`test@example.com` / `password123`
- 管理员用户：`admin@example.com` / `admin123`

### 4. 启动前后端

后端：

```bash
go run ./cmd/picfast
```

前端开发服务器：

```bash
cd web
pnpm install
pnpm dev
```

开发时访问 [http://localhost:5173](http://localhost:5173)，Vite 会代理 `/api`、`/i`、`/t` 到后端 `http://localhost:8080`。

### 5. 本地静态托管前端

如果希望由 Go 服务直接托管前端静态资源：

```bash
cd web && pnpm build
go run ./cmd/picfast
```

服务会优先查找：

1. `PICFAST_SERVER_WEB_DIR` / `server.web_dir`
2. 根目录 `web-dist`
3. 前端默认产物目录 `web/dist`

## 邮箱验证注册

默认情况下，注册成功后会直接登录。只有在以下两个条件同时满足时，系统才会启用“必须验证邮箱后登录”的流程：

1. `app.require_email_verification=true`
2. `mail.*` 已完整配置且 SMTP 可用

对于正式发布，我建议把 `PICFAST_APP_REQUIRE_EMAIL_VERIFICATION` 默认设为 `true`，然后只在填好 SMTP 后再开放注册。

本地开发/联调最省事的方案是直接使用 `Mailpit`：

- SMTP 地址：`mailpit:1025`（Docker 内）或 `127.0.0.1:1025`（本机）
- 邮件预览界面：`http://127.0.0.1:8025`
- 不会真的把邮件发到外部邮箱，只会拦截在本地收件箱里

最小可用配置示例：

```yaml
mail:
  host: "smtp.example.com"
  port: 587
  username: "noreply@example.com"
  password: "your-smtp-password"
  from_email: "noreply@example.com"
  from_name: "PicFast"
  encryption: "starttls"

app:
  allow_registration: true
  require_email_verification: true
```

启用后行为如下：

- 用户注册后不会直接登录，而是收到验证邮件
- 访问邮件中的 `/verify-email?token=...` 链接后即可完成验证
- 未验证邮箱的用户无法登录，登录页可重新发送验证邮件
- 如果只开启 `require_email_verification`，但没有正确配置 SMTP，系统会自动退回普通注册流程，并在启动日志中给出警告
- Docker Compose 示例已经预留了 `PICFAST_MAIL_*` 和 `PICFAST_APP_REQUIRE_EMAIL_VERIFICATION=true`，填入真实邮件参数后即可直接启用
- 当前 Docker Compose 也内置了 `Mailpit`，本地启动后可以直接测试验证邮件流程

建议：

- `server.base_url` 一定要写成用户实际访问的域名，邮件里的验证链接会基于它生成
- 生产环境务必修改 `jwt.secret`
- 可以先只开启 `allow_registration` 做内测，等 SMTP 打通后再启用邮箱验证

### 用 Mailpit 本地测试

```bash
docker compose -f docker/docker-compose.yml up -d db mailpit
go run ./cmd/picfast
```

然后：

1. 打开 `http://127.0.0.1:5173/register`
2. 注册一个新账号
3. 到 `http://127.0.0.1:8025` 查看验证邮件
4. 点击邮件里的验证链接完成验证

## Docker 发布

PicFast 现在可以直接发布到 Docker Hub。当前推荐镜像仓库：

- `xbeta/picfast`

常见标签约定：

- `xbeta/picfast:latest`：`main` 分支最新稳定构建
- `xbeta/picfast:vX.Y.Z`：固定版本发布标签
- `xbeta/picfast:sha-<commit>`：按提交追踪的构建标签

### 3 分钟最小部署（本地先跑起来）

如果你只想快速验证可用性，可以按下面步骤先在本机启动：

```bash
# 0) 创建独立网络（只需一次）
docker network create picfast-net

# 1) 启动 PostgreSQL
docker run -d \
  --name picfast-db \
  --network picfast-net \
  -e POSTGRES_USER=picfast \
  -e POSTGRES_PASSWORD=picfast \
  -e POSTGRES_DB=picfast \
  -v picfast-pgdata:/var/lib/postgresql/data \
  postgres:16-alpine

# 2) 启动 PicFast
docker run -d \
  --name picfast \
  --network picfast-net \
  -p 18080:8080 \
  -e PICFAST_DATABASE_URL='postgres://picfast:picfast@picfast-db:5432/picfast?sslmode=disable' \
  -e PICFAST_JWT_SECRET='replace-with-a-strong-secret' \
  -e PICFAST_SERVER_BASE_URL='http://localhost:18080' \
  -v picfast-uploads:/app/data/uploads \
  -v picfast-thumbnails:/app/data/thumbnails \
  xbeta/picfast:latest
```

启动后访问 `http://localhost:18080`，按页面引导创建第一个管理员账号。线上建议保持应用监听容器内 `8080`，由反向代理统一处理域名与 HTTPS。

清理（删除容器 + 数据卷 + 网络）：

```bash
docker rm -f picfast picfast-db 2>/dev/null || true
docker volume rm picfast-pgdata picfast-uploads picfast-thumbnails 2>/dev/null || true
docker network rm picfast-net 2>/dev/null || true
```

### 拉取镜像

```bash
docker pull xbeta/picfast:latest
```

### 运行示例

```bash
# 如果使用你自己的 PostgreSQL，请将 HOST/USER/PASSWORD/DB 替换为实际值
docker run -d \
  --name picfast \
  -p 18080:8080 \
  -e PICFAST_DATABASE_URL='postgres://USER:PASSWORD@HOST:5432/DB?sslmode=disable' \
  -e PICFAST_JWT_SECRET='replace-with-a-strong-secret' \
  -e PICFAST_SERVER_BASE_URL='https://picfast.example.com' \
  -v picfast-uploads:/app/data/uploads \
  -v picfast-thumbnails:/app/data/thumbnails \
  xbeta/picfast:latest
```

默认情况下，空数据库首次访问会进入初始化向导，由浏览器创建第一个管理员账号。无人值守部署可以额外设置 `PICFAST_APP_ADMIN_EMAIL` 和 `PICFAST_APP_ADMIN_PASSWORD`，应用会在首次启动时自动创建管理员并跳过向导。

### GitHub Actions 自动发布

仓库已包含 Docker 发布工作流 [`docker-publish.yml`](.github/workflows/docker-publish.yml)。

版本发布由 [`release-please.yml`](.github/workflows/release-please.yml) 驱动：当 `main` 上累计了符合 Conventional Commits 的变更后，会自动创建或更新 Release PR；合并后自动打 `v*` tag 并创建 GitHub Release。

触发规则：

- push 到 `main`：发布 `latest`、`main`、`sha-*`
- push `v*` tag：发布对应 semver 标签并刷新 `latest`
- 支持手动触发

在 GitHub 仓库 Secrets 中配置：

- `DOCKERHUB_USERNAME=xbeta`
- `DOCKERHUB_TOKEN=<Docker Hub Access Token>`

建议使用 Docker Hub Access Token，不要直接使用账户密码。

安全发布能力：

- CI 会执行 Trivy 文件系统扫描并上传到 GitHub Security
- Docker 发布会执行 Trivy 镜像扫描（当前拦截 `CRITICAL` 漏洞）
- Docker 发布后会使用 Cosign keyless 对镜像摘要签名（依赖 GitHub OIDC）

### 如何选择 Compose 模板

仓库当前提供两套 Compose，面向不同场景：

- `docker/docker-compose.yml`：本地开发与联调，包含 `mailpit`，默认暴露本机端口，适合快速启动
- `docker/docker-compose.traefik.yml`：生产/私有化部署模板，假设你已有 Traefik 入口，应用通过 label 接入反向代理

如果你不使用 Traefik，也可以继续使用镜像部署 PicFast，只需在外部反代（Nginx/Caddy/NPM/HAProxy 等）把 `443/80` 转发到 PicFast 容器 `8080`，并正确设置 `PICFAST_SERVER_BASE_URL`。

### 生产环境必改项（上线前检查）

至少确认以下配置已经替换为真实值：

- `PICFAST_JWT_SECRET`：使用高强度随机密钥
- `PICFAST_SERVER_BASE_URL`：设置为真实访问域名（如 `https://img.example.com`）
- `POSTGRES_PASSWORD`：避免示例密码

建议同时检查：

- 反向代理正确透传 `Host` 和 `X-Forwarded-Proto`
- 上传大小限制与超时配置符合实际需求
- 若启用邮箱验证，确认 `PICFAST_MAIL_*` 可用且可达

### Traefik 单机模板

如果你已经有现成的 Traefik 反向代理，可以直接使用：

- [`docker/docker-compose.traefik.yml`](docker/docker-compose.traefik.yml)
- [`docker/.env.traefik.example`](docker/.env.traefik.example)

推荐步骤：

```bash
cd docker
cp .env.traefik.example .env
docker compose -f docker-compose.traefik.yml up -d
```

这份模板适合：

- PostgreSQL 与 PicFast 同机部署
- Traefik 通过 Docker label 暴露站点
- 本地存储上传文件与缩略图
- 通过浏览器初始化向导创建第一个管理员账号

当前版本的关键行为：**空数据库首次启动时，应用会自动补齐默认分组、游客分组、本地存储策略；首次访问页面会进入初始化向导创建管理员账号。**

这意味着用户不需要再额外执行一次手工 seed，生产 compose 第一次拉起即可在浏览器完成初始化。若需要无人值守部署，也可以在 `.env` 中设置 `PICFAST_APP_ADMIN_EMAIL/PASSWORD` 自动创建管理员并跳过向导。

## 常用命令

```bash
# 开发
make dev
make seed

# 构建
make build
make frontend
make build-full

# 数据库
make migrate-up    # 可选：手动执行迁移，需要 golang-migrate CLI
make migrate-down  # 可选：手动回滚迁移，需要 golang-migrate CLI
make generate

# 质量
make test
make lint
make format

# Docker
make docker-up
make docker-down
make docker-logs
```

## 目录结构

```text
.
├── cmd/                 # 服务入口与开发 seed 脚本
├── internal/
│   ├── config/          # 配置加载
│   ├── domain/          # 领域模型与常量
│   ├── handler/         # HTTP handlers 与 middleware
│   ├── router/          # 路由注册
│   ├── service/         # 上传、删除、审核、缩略图、存储等业务逻辑
│   ├── sqlc/            # sqlc 生成代码
│   └── testutil/        # 测试数据库与测试辅助
├── migrations/          # 数据库迁移
├── api/                 # OpenAPI 描述
├── docker/              # Dockerfile 与 compose
├── web/                 # React 前端
└── web-dist/            # 可选的根级静态构建产物目录
```

## API 概览

| 端点 | 说明 |
|------|------|
| `POST /api/v1/auth/register` | 用户注册 |
| `POST /api/v1/auth/login` | 用户登录 |
| `POST /api/v1/auth/refresh` | 刷新 Access Token |
| `POST /api/v1/upload` | 游客上传 |
| `POST /api/v1/images` | 认证用户上传 |
| `GET /api/v1/images` | 图片列表 |
| `GET /api/v1/albums` | 相册列表 |
| `GET /api/v1/sharex/config` | 下载 ShareX 配置 |
| `GET /i/{key}.{ext}` | 访问图片 |
| `GET /t/{hash}.png` | 访问缩略图 |
| `GET /health` | 健康检查 |

管理员端点前缀：`/api/v1/admin/*`

## MCP API（稳定字段约定）

PicFast MCP Server 推荐通过 `Authorization: Bearer <API_TOKEN>` 访问，当前已提供稳定的工具响应约定：

- `upload_image` 成功返回 JSON 文本，核心字段包含：
  - `key`, `url`, `markdown`, `html`, `bbcode`
  - `mimetype`
  - `original_size`（上传前字节数）
  - `stored_size`（最终落盘字节数）
  - `processed`（是否经过压缩/转码/水印等处理）
- 工具调用失败时，统一返回 JSON 错误对象：
  - `{"error":{"code":"<ERROR_CODE>","message":"<DETAIL>"}}`
- 当前错误码最小集合：
  - `UNAUTHORIZED`
  - `FORBIDDEN_SCOPE`
  - `INVALID_IMAGE_DATA`
  - `UPLOAD_FAILED`
  - `IMAGE_NOT_FOUND`
  - `INTERNAL_ERROR`

完整 MCP 字段示例、错误码语义与兼容建议请见：[`docs/mcp-api.md`](docs/mcp-api.md)。

## 测试与校验

```bash
make test
make lint
```

说明：

- `go test ./...` 会在默认本地 Postgres 上自动创建 `picfast_test` 测试库（如不存在）。
- `make lint` 会执行 `go vet`、前端 ESLint 和 TypeScript 编译检查。

## 部署说明

- Docker 镜像会在构建时编译前端，并通过 `PICFAST_SERVER_WEB_DIR=/web-dist` 交给 Go 服务托管。
- `docker/docker-compose.traefik.yml` 会通过 `init-permissions` 自动创建并修正 `./data/uploads`、`./data/thumbnails` 的写入权限，应用容器本身仍以非 root 用户运行。
- `/docs` 会加载 OpenAPI 文档页，`/openapi.yaml` 提供规范文件下载。
- `/api/v1/admin/debug/pprof/*` 默认关闭；调试时可设置 `PICFAST_SERVER_PPROF_ENABLED=true`，启用后仍需管理员权限访问。
- 如果要在 Docker / 服务器中启用邮箱验证，记得同时传入 `PICFAST_SERVER_BASE_URL`、`PICFAST_MAIL_*` 和 `PICFAST_APP_REQUIRE_EMAIL_VERIFICATION=true`。
