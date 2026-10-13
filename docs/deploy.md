# 反向代理部署

PicFast 后端依赖请求头中的 `X-Forwarded-For` / `X-Real-IP` 来获取真实客户端 IP，用于限流、审计日志、注册记录等功能。

## 必须设置的反向代理头

请确保你的反向代理（nginx、Caddy、Traefik、SafeLine/雷池等）至少传递以下请求头：

### nginx

```nginx
location / {
    proxy_pass http://picfast_backend:8080;

    # 以下 4 个 header 为必须
    proxy_set_header Host $http_host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### Caddy

```caddy
example.com {
    reverse_proxy picfast_backend:8080 {
        header_up Host {http.request.host}
        header_up X-Real-IP {http.request.remote}
        header_up X-Forwarded-For {http.request.remote}
        header_up X-Forwarded-Proto {http.request.scheme}
    }
}
```

### Traefik

Traefik 默认会将 `X-Forwarded-For`、`X-Real-IP` 等头设为受信头，无需额外配置。
如需显式声明，可以在入口点或路由上启用：

```yaml
entryPoints:
  websecure:
    address: ":443"
    forwardedHeaders:
      trustedIPs:
        - "10.0.0.0/8"
        - "172.16.0.0/12"
```

### SafeLine / 雷池

在站点配置中，确保以下 header 已启用（默认通常已启用 `X-Forwarded-For`）：

- `Host`
- `X-Forwarded-For`
- `X-Forwarded-Proto`

如果雷池内置 nginx 未自动添加 `X-Forwarded-For`，可在「高级配置」中手动添加：

```nginx
proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
proxy_set_header X-Real-IP $remote_addr;
```

## 可信代理配置

为防止客户端伪造 IP（IP Spoofing），PicFast 支持配置可信代理 CIDR 列表。
仅在 `trusted_proxies` 中的来源 IP 会被信任，从而读取其转发的 `X-Forwarded-For` 头。

在 `config.yaml` 中配置：

```yaml
server:
  trusted_proxies:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
    - "192.168.0.0/16"
```

或通过环境变量：

```bash
export PICFAST_SERVER_TRUSTED_PROXIES="10.0.0.0/8,172.16.0.0/12"
```

### 工作原理

1. 请求到达时，PicFast 在 RealIP 中间件之前，将握手时的 TCP 来源地址（`r.RemoteAddr`）保存到请求上下文
2. 若该原始地址属于 `trusted_proxies` 中的 CIDR，启用 `RealIP` 中间件从 `X-Forwarded-For` / `X-Real-IP` 头中解析真实客户端 IP
3. 若原始地址**不**在可信代理列表中，忽略所有代理头，防止客户端伪造 IP
4. 后续业务代码通过 `clientip.FromRequest(r)` 统一获取真实客户端 IP（基于上下文中的原始地址判定受信关系）

**注意**：未配置 `trusted_proxies` 时，`RealIP` 中间件不会生效，`X-Forwarded-For` 头将被忽略。

## 镜像变体

- **`latest` / `vX.Y.Z`（完整镜像）**：内置 libvips，支持 WebP 输出、文字水印、以及基于 `pg_dump` 的维护备份/恢复。
- **`slim` / `vX.Y.Z-slim`（精简镜像）**：纯 Go 构建（`CGO_ENABLED=0`，无 libvips），镜像更小、冷启动更快，推荐用于容器 PaaS。不支持 WebP 输出与文字水印，也不含 `pg_dump` 维护备份/恢复；注意即使未开启“去除 EXIF”，重新编码的图片也不会保留 EXIF 元数据。

## 健康检查端点

PicFast 在服务根路径暴露两个探针，适用于负载均衡器与容器平台的健康检查：

- `GET /healthz` — 存活探针，始终返回 200，不访问数据库或磁盘
- `GET /readyz` — 就绪探针，仅做数据库连通性检查，失败返回 503

如果反向代理只转发 `/api` 前缀，请确认这两个路径可达（或单独配置转发规则）。

## 数据库迁移

数据库迁移已嵌入二进制，服务启动时自动执行。容器 PaaS / 多副本部署时如需统一控制迁移时机：

```bash
export PICFAST_DATABASE_AUTO_MIGRATE=false
# 在启动服务前单独执行（同一镜像/二进制）
picfast migrate
```

## 托管 Postgres 与连接池

使用事务模式（transaction mode）连接池的托管 Postgres（Supabase 的 `:6543` 端口、Neon 的 `-pooler` 主机、pgbouncer transaction pooling）时，pgx 默认的语句缓存在池化连接上不可靠 —— 在 `database.url` / `PICFAST_DATABASE_URL` 的查询串追加 `default_query_exec_mode=exec` 即可：

```
postgres://user:pass@host:6543/db?sslmode=require&default_query_exec_mode=exec
```

也可以改用直连/会话模式的连接串。后台任务的互斥锁使用事务级咨询锁，两种模式下均能正常工作。

## 后台任务（webhook 投递与清理）

PicFast 的后台任务包含 webhook 投递（outbox → 各订阅端点）与每小时清理（过期图片、webhook 投递日志、审计日志）。默认在进程内按 `webhook.worker_interval`（webhook）与每小时（清理）自动运行，适用于常驻部署。

在无常驻进程的平台（serverless / 容器实例会在响应后回收）上可关闭内置调度，改为外部定时触发：

```bash
export PICFAST_JOBS_EMBEDDED=false   # 关闭进程内调度
export PICFAST_JOBS_TOKEN=<随机长令牌>   # 启用 POST /internal/jobs/run（Bearer 鉴权；不设置则不注册该端点）
```

两种触发方式等价，按需任选：

- **HTTP**：`POST /internal/jobs/run`，请求头 `Authorization: Bearer $PICFAST_JOBS_TOKEN`，同步执行一轮并返回 JSON 报告（任务失败返回 500）
- **CLI**：`picfast jobs run`（同一镜像/二进制），同步执行一轮并把报告打印为 JSON，失败时退出码为 1

外部调度示例：

```yaml
# Google Cloud Scheduler —— HTTP 触发，建议每分钟一次
gcloud scheduler jobs create http picfast-jobs \
  --schedule="* * * * *" \
  --uri="https://<host>/internal/jobs/run" \
  --http-method=POST \
  --headers="Authorization=Bearer $PICFAST_JOBS_TOKEN"
```

```yaml
# GitHub Actions —— .github/workflows/jobs.yml
on:
  schedule:
    - cron: "* * * * *"   # GitHub 最短间隔约 5 分钟
jobs:
  run:
    runs-on: ubuntu-latest
    steps:
      - run: curl -fsS -X POST -H "Authorization: Bearer ${{ secrets.PICFAST_JOBS_TOKEN }}" https://<host>/internal/jobs/run
```

```bash
# Fly.io —— 用同一镜像定时执行 CLI
fly machine run <image> --schedule hourly picfast jobs run
```

```bash
# Railway —— cron 服务直接执行
picfast jobs run
```

注意：

- `embedded=false` 时 webhook 投递延迟取决于外部调度频率，建议每分钟触发一次
- 多实例并发触发是安全的：各任务通过 Postgres 咨询锁串行执行，重复触发只会标记为 `skipped`
- 限流器是按实例生效的：多副本部署时每实例独立计数，实际限额约等于配置值 × 实例数

## 站点设置同步

管理后台保存的站点设置会写入数据库；多副本部署时其他实例会在约 10 秒内自动应用最新设置，无需重启。`moderation_mode` 同样在切换后的下一个请求即生效，无需重启。

## 缩略图存储

缩略图随图片保存在同一存储策略的 `_picfast/thumbnails/` 对象路径下，`/t/{md5}.png` 仍由应用统一提供（缺失时会从原图惰性重建并回写存储）。`storage.thumbnail_dir` 现在只是只读的旧版回退目录，仅用于继续提供升级前生成的磁盘缩略图，服务不会向其中写入新文件。

将旧磁盘缩略图迁移到存储策略：

```bash
picfast maintenance repair-thumbnails --apply
```

## 默认存储策略（首次启动）

当数据库中尚无任何存储策略时，可用环境变量指定首个策略的类型与配置：

```bash
export PICFAST_STORAGE_DEFAULT_TYPE=s3
export PICFAST_STORAGE_DEFAULT_CONFIG='{"endpoint":"https://s3.example.com","region":"auto","bucket":"pics","access_key":"AK","secret_key":"SK","url":"https://cdn.example.com"}'
```

Cloudflare R2（S3 兼容）示例：

```bash
export PICFAST_STORAGE_DEFAULT_TYPE=s3
export PICFAST_STORAGE_DEFAULT_CONFIG='{"endpoint":"https://<account>.r2.cloudflarestorage.com","region":"auto","bucket":"pics","access_key":"AK","secret_key":"SK","use_path_style":true}'
```

配置仅在没有现存策略时生效；策略已存在则忽略并记录日志。也可以在 `config.yaml` 中以 YAML map 形式书写 `storage.default_config`。

> ⚠️ 容器 PaaS（无持久卷）上默认的 `local` 存储位于临时磁盘，实例重建后上传文件即丢失 — 部署到此类平台时务必配置持久后端。
