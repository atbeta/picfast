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
