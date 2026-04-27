# PicFast

PicFast 是一个现代化的图床/图片托管服务，支持多用户、多存储策略、分组权限管理和 SAAS 化部署。

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.26, Chi Router, pgx/v5, sqlc, JWT |
| 前端 | Vue 3, TypeScript, Vite, Naive UI, Tailwind CSS v4 |
| 数据库 | PostgreSQL 16 |
| 存储 | 本地文件系统 / AWS S3 兼容对象存储 |
| 监控 | Prometheus (metrics 端点) |

## 快速开始

### 前置要求

- Go 1.26+
- Node.js 20+ + pnpm
- PostgreSQL 16 (或 Docker)
- `golang-migrate` CLI (数据库迁移)

### 1. 启动数据库

```bash
# 使用 Docker（推荐）
make docker-up

# 或使用本地 PostgreSQL
createdb picfast
```

### 2. 配置环境

```bash
cp .env.example .env
# 按需编辑 .env 文件
```

### 3. 运行数据库迁移

```bash
make migrate-up
```

### 4. 填充开发数据

```bash
make seed
```

这会创建：
- 默认用户组 + 游客组
- 本地存储策略
- 测试用户：`test@example.com` / `password123`
- 管理员用户：`admin@example.com` / `admin123`

### 5. 启动后端

```bash
make run
# 或开发模式（热重载需自行配置 air）
go run ./cmd/picfast
```

### 6. 启动前端

```bash
cd web && pnpm install && pnpm dev
```

前端开发服务器会代理 API 请求到 `http://localhost:8080`。

访问 http://localhost:5173 即可使用。

## 目录结构

```
.
├── cmd/
│   ├── picfast/          # 主服务入口
│   └── seed/            # 开发数据填充脚本
├── internal/
│   ├── config/          # 配置管理 (Viper)
│   ├── domain/          # 领域类型和常量
│   ├── handler/         # HTTP handlers
│   │   └── middleware/  # 认证、限流、日志、指标
│   ├── router/          # 路由注册
│   ├── service/         # 业务逻辑 (上传、删除、水印、缩略图)
│   │   └── storage/     # 存储抽象 (Local / S3)
│   ├── sqlc/            # sqlc 生成的类型安全数据库代码
│   │   └── queries/     # SQL 查询定义
│   └── testutil/        # 测试辅助工具
├── web/                 # Vue 3 前端
│   └── src/
│       ├── api/         # API 客户端和类型
│       ├── views/       # 页面组件
│       └── stores/      # Pinia 状态管理
├── migrations/          # 数据库迁移文件
├── docker/              # Docker 配置
└── docs/                # 文档
```

## 常用命令

```bash
# 开发
make dev          # 显示开发环境启动指南
make seed         # 填充开发测试数据

# 构建
make build        # 编译后端
make build-full   # 编译前端 + 后端
make frontend     # 仅编译前端

# 数据库
make migrate-up   # 执行迁移
make migrate-down # 回滚一步
make generate     # 重新生成 sqlc 代码

# 质量
make test         # 运行全部测试
make format       # 格式化 Go + 前端代码
make lint         # 静态检查
make tidy         # 整理 Go 依赖

# Docker
make docker-up    # 启动 Docker 环境
make docker-down  # 停止 Docker 环境
make docker-logs  # 查看应用日志

# 其他
make clean        # 清理构建产物
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
| `GET /i/{key}.{ext}` | 访问图片 |
| `GET /t/{hash}.png` | 访问缩略图 |
| `GET /health` | 健康检查 (含数据库/存储状态) |
| `GET /metrics` | Prometheus 指标 |

管理员端点前缀：`/api/v1/admin/*`
pprof 调试端点：`/api/v1/admin/debug/pprof/*`

### API 快速调用示例

**上传图片（curl）**
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "file=@photo.png" \
  -F "expires_in=24h"
```

**上传图片（TypeScript / Fetch）**
```typescript
const file = document.getElementById('file').files[0];
const form = new FormData();
form.append('file', file);
form.append('expires_in', '24h');

const res = await fetch('/api/v1/upload', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer <token>', // optional for guest upload
  },
  body: form,
});
const data = await res.json();
console.log(data.data.url);          // https://...
console.log(data.data.markdown);     // ![photo.png](https://...)
console.log(data.data.thumbnail_url);// https://.../t/xxx.png
```

**ShareX 配置下载**
```bash
curl http://localhost:8080/api/v1/sharex/config \
  -o picfast.sxcu
```

## 核心功能

- **多存储策略**：支持本地磁盘和 S3 兼容对象存储，可按用户组分配
- **图片处理**：格式转换、质量压缩、文字水印（5 种位置 + 透明度）
- **文件去重**：MD5 + SHA1 双哈希去重，节省存储空间
- **缩略图**：异步生成 400px 缩略图
- **权限控制**：公开/私有图片，私有图片需所有者才能访问
- **速率限制**：按分钟/小时/天/月多维度限流
- **容量配额**：用户级别存储容量限制
- **JWT 认证**：Access Token + Refresh Token 双令牌机制

## 配置

应用支持通过 `config.yaml` 或环境变量配置。环境变量前缀为 `PICFAST_`，层级用 `_` 分隔。

示例：
```yaml
server:
  port: 8080
  base_url: "http://localhost:8080"

database:
  url: "postgres://picfast:picfast@localhost:5432/picfast?sslmode=disable"

jwt:
  secret: "change-me-in-production"
  access_ttl: 15m
  refresh_ttl: 168h

storage:
  local_root: "./data/uploads"
  thumbnail_dir: "./data/thumbnails"

app:
  name: "PicFast"
  allow_guest_upload: true
  allow_registration: true
  user_initial_capacity: 524288000
  admin_email: ""
  admin_password: ""
```

## 测试

```bash
# 全部测试
make test

# 仅后端服务测试
go test -v ./internal/service/...

# 仅中间件测试
go test -v ./internal/handler/middleware/...

# 启动测试数据库容器
make docker-test-db
```

## 部署

```bash
# Docker Compose
make docker-up

# 手动构建
make build-full
./bin/picfast
```

生产环境部署前请务必：
1. 修改 JWT Secret
2. 配置 HTTPS Base URL
3. 设置强密码的 Admin 账户
4. 配置 S3 存储或使用持久化的本地卷

## 致谢

PicFast 在功能设计和交互理念上受到了 [lsky-pro](https://github.com/lsky-org/lsky-pro) 的启发。
本项目使用全新的技术栈（Go + Vue3）独立实现，谨向 lsky-pro 团队对开源图床领域的贡献致以敬意。

## License

本项目采用 [GNU General Public License v3.0](LICENSE) 协议开源。

> **非商业声明**：PicFast 为社区驱动项目，不计划商业化运营，仅接受捐赠以维持服务器和开发成本。
> 你可以自由使用、修改和分发，但请遵守 GPL v3 的 copyleft 要求。
