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
