# Serverless / 容器 PaaS 部署

0.26 起 PicFast 提供了纯 Go 的 **slim 镜像**（无 libvips/CGO、distroless、体积约为完整镜像的 1/10），同时补齐了"无常驻实例"平台需要的接口：`PORT` 监听、`/healthz`+`/readyz` 探针、内嵌迁移与 `picfast migrate`、外部触发后台任务、首启存储策略环境变量。

本文给出在常见 serverless / 容器 PaaS 上部署可用的 PicFast 的方式，示例平台：**Fly.io、Railway、Render、Koyeb、Zeabur**。反向代理（nginx/Caddy/Traefik）与自建 Docker Compose 部署见 [deploy.md](deploy.md) 与 [../docker/README.md](../docker/README.md)。

> **一句话结论**：这类平台上的实例磁盘都是临时的、实例可能随时被回收，所以 PicFast 必须凑齐"三件套"——**外部 PostgreSQL**、**S3 兼容对象存储**、**外部定时触发**（用于 webhook 投递与每小时清理）。

---

## 1. 这类平台的三个前置条件

### 1.1 外部 PostgreSQL

平台只跑 PicFast 进程，数据库用托管 PG（Neon、Supabase、RDS、自建等均可）：

```bash
PICFAST_DATABASE_URL=postgres://user:pass@host:5432/picfast?sslmode=require
```

- **免费且不需要绑卡的选项**：Neon（永久免费、无需绑卡：1 GB 存储 + 100 CU-小时/月、空闲 5 分钟缩容到零，官方明确 **Free 计划不对超额计费**，不会意外扣款）、Aiven（free forever、无需绑卡）、Prisma Postgres（free forever、无需绑卡）。注意 Render 自带的免费 Postgres **创建 30 天后过期**，不建议长期使用。
- **免费额度会被后台任务吃掉**：内置后台任务每 5 秒轮询一次数据库（`webhook.worker_interval`），只要实例在线，Neon 这类"空闲即缩容"的数据库就不会休眠——长期在线会消耗掉每月的计算额度。若因此超额，请设 `PICFAST_JOBS_EMBEDDED=false` 并改用外部定时触发（见 §1.3）。
- **事务模式连接池**（Supabase `:6543`、Neon `-pooler`、pgbouncer）需在查询串追加 `default_query_exec_mode=exec`，或直接使用直连/会话模式串；pgx 的默认语句缓存在事务池化下不可靠。迁移也必须用直连串（见 §7）。
- 保持 `pool_max_conns >= 3`：后台任务持锁时会占用一条连接，另需连接执行查询；过小会在启动时打印警告。
- **PicFast 启动即连库**，连不上或 `pool.Ping` 失败会直接退出。如果托管数据库会在空闲后自动挂起（部分免费档），缩容到零后的第一次请求可能因为同时唤醒数据库而失败一次，建议把数据库保持在常驻状态，或给服务设置最小实例数 ≥ 1。

### 1.2 持久对象存储（不上 S3 就一定会丢图）

这些平台的容器文件系统都是**临时**的：重启、重新部署、实例被回收都会丢数据，而且**没有平台允许把同一块卷共享给多个实例**，因此不能靠本地卷跑多副本。请使用 S3 兼容存储（R2 / TOS / OBS / Kodo / OSS / COS / MinIO 均可）。

- **免费且不需要绑卡的选项**：Backblaze B2（前 10 GB 永久免费）、Tigris（5 GB）、Neon Object Storage（同一 Neon 账号内 5 GB；仅支持 path-style 与 SigV4，bucket 的公开读要在 Neon 控制台设置，否则把存储策略的 `link_mode` 设为 `proxy` 交给 PicFast 出图）。
- ⚠️ **Cloudflare R2 需要绑定支付方式**才能开通，若坚持不绑卡请避开它。

首次启动（数据库里还没有任何存储策略时）用环境变量创建默认策略：

```bash
# S3 / MinIO
PICFAST_STORAGE_DEFAULT_TYPE=s3
PICFAST_STORAGE_DEFAULT_CONFIG='{"endpoint":"https://s3.example.com","region":"auto","bucket":"pics","access_key":"AK","secret_key":"SK","url":"https://cdn.example.com"}'

# Cloudflare R2（S3 兼容）
PICFAST_STORAGE_DEFAULT_TYPE=s3
PICFAST_STORAGE_DEFAULT_CONFIG='{"endpoint":"https://<account>.r2.cloudflarestorage.com","region":"auto","bucket":"pics","access_key":"AK","secret_key":"SK","use_path_style":true}'
```

> 该变量**只在数据库尚无任何存储策略时生效**（首启）；策略已存在时会被忽略并打印 WARN。此时：
>
> - **推荐**：到管理后台「存储策略」新增 S3/R2 策略并设为分组默认（旧策略与历史图片不受影响）；
> - 若这个库还没上传过图片、想继续走环境变量：`delete from strategies;` 后重启——**注意副作用**：`group_strategies` 会级联删除，`images.strategy_id` 会被置为 `NULL`，所以有历史图片时不要这么做；
> - 或者把现有行**就地改造**（保留分组映射与图片引用）：`update strategies set name='Default Storage', strategy_type='s3', configs='{...}'::jsonb where id = 1;`
>
> ⚠️ **留空与写错的后果不同**：只设 `PICFAST_STORAGE_DEFAULT_TYPE`（例如 `s3`）而把 `PICFAST_STORAGE_DEFAULT_CONFIG` 留空，应用会**回退到本地存储并打印 WARN**（这样 PaaS 的创建页可以把它做成“可填可不填”的提示项）；而 JSON **写错**时会拒绝启动（刻意 fail-fast，避免把打错的配置静默当作“用本地盘”从而丢图）。

### 1.3 外部定时触发（webhook 与清理）

常驻部署下，webhook 投递与每小时清理由进程内调度器负责。实例会在响应后被回收的平台上应关闭内置调度，改由外部定时触发：

```bash
PICFAST_JOBS_EMBEDDED=false
PICFAST_JOBS_TOKEN=<随机长令牌>   # 不设置则不注册 HTTP 端点
```

两种等价的触发方式，按平台能力任选：

- **HTTP**：`POST /internal/jobs/run`，请求头 `Authorization: Bearer $PICFAST_JOBS_TOKEN`，同步执行一轮并返回 JSON 报告（任务失败返回 500）。该端点按来源 IP 限流（10 次/分钟）。
- **CLI**：`picfast jobs run`（同一镜像），同步执行一轮并打印 JSON，失败退出码 1。

建议**每分钟触发一次**（webhook 投递延迟 = 触发间隔）。多实例并发触发是安全的：任务通过 Postgres 咨询锁串行执行，重复触发只会标记 `skipped`。

---

## 2. 镜像选择：slim 还是完整镜像

| 标签 | 说明 |
|------|------|
| `latest` / `v0.26.0` | **完整镜像**（debian-slim + libvips + `pg_dump`）：WebP 输出、文字水印、维护备份/恢复。体积大、冷启动慢。 |
| `slim` / `v0.26.0-slim` / `0.26-slim` | **精简镜像**（纯 Go、distroless、无 CGO）：amd64 压缩后约 13 MB（完整镜像约 120 MB），冷启动快，推荐 serverless / 容器 PaaS 使用。 |

slim 的能力差异（部署前请确认可接受）：

- 不支持 **WebP 输出**与**文字水印**（前端会自动隐藏服务端无法应用的选项）；
- 不含 `pg_dump`，管理后台的**维护备份/恢复不可用**（`picfast maintenance backup` 依赖 `pg_dump`）——需要内置备份能力时请用完整镜像，或在外部用 `pg_dump` / 托管数据库自带的快照；
- 重新编码不保留 EXIF 元数据（即使未开启"去除 EXIF"）；
- 无法对 **WebP 源图**做动态缩放/压缩——对 `.webp` 请求 `@w_`/`@q_` 变体会返回原图（WebP 缩略图仍可生成）。

> ⚠️ **slim 是 distroless 镜像，容器里没有 shell**（也没有包管理器）。因此平台侧凡是"通过 shell 执行命令"的地方都会失败：`sh -c '...'`、`bash -c`、`&&`、管道、重定向。能用的方式是**让平台直接 exec 二进制**，即"覆盖 Docker `CMD`、参数追加到镜像 `ENTRYPOINT`"这种语义——此时迁移命令填 `migrate`、后台任务填 `jobs run`。
>
> 平台把命令当作**完整命令行**执行时，则要写完整路径 `/usr/local/bin/picfast migrate`（slim 与完整镜像的二进制都在这个位置）。两种写法只差一个前缀，判断方法很简单：**看平台部署日志**，出现 `unknown command "/usr/local/bin/picfast"` 说明平台走的是 CMD 语义（去掉路径），出现 `executable file not found` 说明要补上完整路径。`sh -c "..."`、`bash -c "..."`、`cmd1 && cmd2` 这类写法在 slim 上一定失败（Render 文档就建议用 `/bin/sh -c` 串多条命令——slim 不支持，需要改用完整镜像或拆成单条命令）。

**实例内存建议**：slim 使用标准库解码器（整张位图都在内存里），峰值约为 `3 × 宽 × 高 × 4` 字节（单张图的解码像素上限为 6400 万，最坏情况约 750 MB），并发解码上限为 2。开启图片处理时建议单实例 **≥ 1 GB 内存**；如果平台只给 256–512 MB，请关闭图片处理或下调上传上限：

```bash
PICFAST_APP_SKIP_IMAGE_PROCESSING=true     # 也可在后台"站点设置"里关闭（后台设置优先于环境变量）
PICFAST_APP_MAX_UPLOAD_BYTES=10485760      # 默认 52428800（50 MiB）
```

PicFast 自身的请求时限：读取请求 30s、写响应 60s，另有 60s 的请求超时中间件（可用 `server.read_timeout` 调整）。平台的请求超时设置成 **≥ 60s** 即可，更长没有意义。

> 生产部署建议用**固定版本标签**（如 `v0.26.0-slim`）：`slim`、`0.26-slim` 这类滚动标签会随新版本移动，而多数 PaaS 不会因为镜像 tag 变化自动重新部署（Render、Koyeb 都需要手动触发）。

---

## 3. 通用环境变量（最小可运行集合）

```bash
# 必填
PICFAST_DATABASE_URL=postgres://user:pass@host:5432/picfast?sslmode=require
PICFAST_JWT_SECRET=<强随机密钥>
PICFAST_SERVER_BASE_URL=https://pics.example.com     # 站点对外地址，影响图片 URL 与登录跳转

# 平台差异项：端口
# PICFAST_SERVER_PORT 优先于平台注入的 PORT；两者都没设置时用 8080
PICFAST_SERVER_PORT=8080

# serverless 平台建议
PICFAST_SERVER_METRICS_ENABLED=false                  # 关掉独立的 /metrics 监听（默认 127.0.0.1:9090，PaaS 抓不到）
PICFAST_JOBS_EMBEDDED=false
PICFAST_JOBS_TOKEN=<随机长令牌>
PICFAST_DATABASE_AUTO_MIGRATE=false                  # 迁移交给发布钩子/外部执行（见 §7）

# 首启存储（§1.2）与无人值守初始化
PICFAST_STORAGE_DEFAULT_TYPE=s3
PICFAST_STORAGE_DEFAULT_CONFIG='{...}'
PICFAST_APP_ADMIN_EMAIL=admin@example.com
PICFAST_APP_ADMIN_PASSWORD=<强密码>
```

要点：

- **端口优先级**：`PICFAST_SERVER_PORT` > 平台注入的 `PORT` > 默认 `8080`。平台注入了 `PORT` 时 PicFast 会自动使用；平台不注入时（例如 Fly.io）用默认 8080，或在 `PICFAST_SERVER_PORT` 里显式指定。
- **站点设置会盖住环境变量**：管理后台保存过的站点设置（站点名、站点地址、是否允许注册/游客上传、图片处理开关等）持久化在数据库里，之后环境变量里的 `PICFAST_APP_*`、`PICFAST_SERVER_BASE_URL` 不再生效。首次部署用环境变量引导，之后改配置请到后台改。
- 不设置 `PICFAST_APP_ADMIN_EMAIL`/`PASSWORD` 时，首次访问会进入初始化向导；无人值守部署建议直接用环境变量创建管理员。

健康探针（服务根路径，无需鉴权）：

| 端点 | 用途 |
|------|------|
| `GET /healthz` | 存活探针，恒返回 200，不访问数据库与磁盘 |
| `GET /readyz` | 就绪探针，仅做数据库连通性检查，失败返回 503 |
| `GET /health` | 旧版端点，**还会检查 `storage.local_root` 是否可写**；用它做探针时请确保该目录可写（必要时指向 `/tmp` 下的路径） |

---

## 4. 平台对照表

以各平台官方文档为准（下表为撰写时的核对结果，涉及具体数值时请以平台最新文档为准）：

| | Fly.io | Railway | Render | Koyeb | Zeabur |
|---|---|---|---|---|---|
| 直接部署 Docker Hub 公开镜像 | ✅ `fly launch --image` / `[build] image` | ✅ 控制台选 Docker Image | ✅ Source → Existing Image | ✅ `--docker docker.io/...` | ✅ Add Service → Docker Images |
| 端口 | 不注入 `PORT`，用 `[http_service] internal_port`（默认 8080） | 注入 `PORT`（未公开默认值），也可用 Target Port 自动探测 | 注入 `PORT`，**默认 10000** | 注入 `PORT` = 你暴露的最小端口号 | 镜像服务需自行声明端口；`PORT` 注入情况文档不明确 |
| 健康检查 | `[[http_service.checks]]`（要求 2xx） | `healthcheckPath`（任意 2xx，默认超时 300s） | `healthCheckPath`（2xx/3xx，5s 内响应） | 默认 TCP；可配 HTTP（2xx/3xx） | 默认 TCP；可设 HTTP 路径（2xx） |
| 迁移钩子 | `[deploy] release_command`（覆盖 CMD，ENTRYPOINT 仍会执行） | Pre-deploy Command（独立容器，无卷） | `preDeployCommand` | 无（用 `koyeb services exec`） | 无（用控制台 Command） |
| 定时任务 | `fly machine run --schedule`（最短每小时）；需每分钟用外部调度或 supercronic | 服务级 Cron Schedule，**最短间隔 5 分钟** | Cron Jobs 服务（支持预构建镜像） | 文档中无定时任务功能 | 文档中无定时任务功能 |
| 缩容到零 | ✅ 默认 `auto_stop_machines = "stop"` | ✅ Serverless（空闲 5–10 分钟，首次请求可能 502） | 仅 Free 档（空闲 15 分钟） | ✅ 预览特性（按实例类型，空闲 5 分钟起） | ❌ 只能手动 Suspend |
| 卷能否多实例共享 | ❌ 一卷只能挂一台 Machine | ❌（卷会带来停机） | ❌（挂盘后不能扩容且失去零停机） | ❌（预览，仅 was/fra，1–10 GB，scale 必须为 1） | ❌ 卷不跨服务共享 |
| 平台公开入口代理网段 | 未公开（提供 `Fly-Client-IP`） | 未公开（文档提供 `X-Real-IP`） | 未公开 | 未公开 | 未公开 |

**结论**：不要把上传落在本地卷上。对象存储 + 无状态实例是这五个平台上唯一能稳定扩容的形态。

### 4.1 绑卡与免费额度（2026-10 核对）

| 平台 | 是否需要绑银行卡 | 免费额度 | 备注 |
|------|------------------|----------|------|
| **Render** | ❌ Free 档不需要 | 750 实例小时/月、0.1 CPU / 512 MB | **cron 没有免费档**；免费 Postgres **30 天后过期**；15 分钟无流量休眠、约 1 分钟启动 |
| **Railway** | ❌ "No credit card required" | 30 天 $5 试用 → Free **$1/月** | 原生 cron 最短 5 分钟；睡醒后首个请求可能 502 |
| **Fly.io** | ⚠️ 试用不需要，之后必须 | 2 机器小时或 7 天 | 不绑卡则应用会停跑 |
| **Koyeb** | ✅ 必须绑卡 | 新用户只能订 Pro 及以上，需 $29 预授权 | 免费档文档还在，但新用户不可达 |
| **Northflank** | ✅ 所有计划都要 | Dev Sandbox：2 服务 / 2 任务 / 1 addon | 官方明确要求先添加支付方式 |
| **Zeabur** | ❌ 不需要 | Free 计划是"管理自有服务器" | 共享集群已停止接收新项目，等于没有免费算力 |

**不绑卡的最小可用组合**：**Render Free（容器）+ Neon（PostgreSQL，永久免费）+ Backblaze B2 或 Neon Object Storage（对象存储）**——三项都不需要银行卡，Neon 一家即可同时提供 PG 与 5 GB 对象存储。需要 cron 准点投递或更大内存时，改用 Railway（原生 cron，最短 5 分钟）或 Render 付费档。

### 4.2 中国大陆平台（实名即可，与境外信用卡无关）

国内平台的注册门槛是**实名认证**，付款走微信/支付宝即可；但"零成本试用"要看各家免费额度，且**镜像能否直接用 Docker Hub 往往才是关键**：

| 平台 | 能否直接用公开镜像 | 免费额度 | 主要坑 |
|------|--------------------|----------|--------|
| **腾讯云 CloudBase 云托管** | ✅ **可以**（官方文档明确列 Docker Hub） | 免费体验版 0 元/月 + **3000 资源点/月**（可续 6 个月；【推算】约等于 1C2G 常驻 37.5 小时，或出网流量约 6.5 GB） | "免费体验版能否开通云托管"三份官方文档互相矛盾，**需要实测**；额度是**多产品共享**的，云存储/CDN 流量最贵（460 点/GB）；🔴 官方明文"**不要在请求处理程序范围之外启动后台线程或进程**"，必须关闭内置调度 |
| **阿里云函数计算 FC 3.0** | ❌ 必须用**同账号同地域 ACR** 中的镜像 | 15 万 CU/周期 × 连续 3 个周期 | 超出额度**自动转按量付费**；**不注入 `PORT`**（平台只提供 `FC_CUSTOM_LISTEN_PORT`）——需把"监听端口"（示例 9000）配成 8080，并用 `PICFAST_SERVER_PORT` 显式对齐 |
| **火山引擎 veFaaS** | ❌ 必须推同地域镜像仓库（CR 体验版已下线） | ❌ **无**：新用户试用包 2024-05 已下线 | Web 应用函数默认端口 **8000**，须设 `PICFAST_SERVER_PORT=8000`；后台进程"不鼓励"；零成本不成立 |
| **腾讯云 SCF（容器镜像）** | ❌ 必须推 TCR（**个人版免费**，可 retag 绕过） | 前 3 个月每月 100 万次调用 + 100 万 GBs + 2GB 出流量，第 4 个月起按量 | 端口固定 **9000**、HTTP Server 须 30 秒内就绪；适合做兜底或定时任务，不适合当 Web 主入口 |
| **华为云 FunctionGraph** | ❌ 必须推 SWR（基础版本身免费） | ✅ 长期每月：请求 ≤100 万次、执行 ≤40 万 GB-秒 | 端口写死 **8000**；🔴 **HTTP 函数不支持定时触发器**（后台任务只能靠事件函数） |
| **百度智能云 CFC** | ❌ **没有容器镜像部署入口**（只能传 zip） | 0 元新客包连续 3 个月 | 对"容器化 serverless"直接出局 |

- **技术契合度最高的是 CloudBase 云托管**：唯一能直接拉 Docker Hub 公开镜像的国内平台，原生注入 `PORT`（`PORT` 值始终反映请求到达的端口），请求超时 60s 正好等于 PicFast 的 `WriteTimeout`，而且**平台自带 PostgreSQL 与云存储**（环境内网互通），不用再找外部 PG/对象存储。冷启动官方没有给出量化数字（只说明与资源状态、镜像大小、启动时间有关——网上流传的"30 秒"是腾讯云 SCF 的启动超时，别套用到 CloudBase）。代价是得先确认免费体验版能否创建云托管服务，且免费额度只够功能验证/演示。
- **对象存储的免费额度**：京东云 OSS 是国内唯一**按月持续**发放的（标准存储 10 GB/地域/月，当月有效不累计），而且免费试用专区另有**新人专享「对象存储资源包 50 GB / 0 元 / 6 个月」并每月再送 10 GB**，前半年比腾讯云 COS（一次性 50 GB/180 天）宽裕；阿里云 OSS 是一次性 3 个月 20 GB；火山 TOS / 华为 OBS / 百度 BOS 没有免费额度；青云 QingStor 是长期定价阶梯（存储 0–10 GB、下载 0–1 GB 免费）。注意京东云免费专区里**没有**函数计算、JKE、JCR、RDS，别去那里找容器产品。PG 方面，除 CloudBase 内置库外，阿里云 RDS PostgreSQL 只送 1 个月试用且**到期自动扣费**（需手动释放），PolarDB 的"前三个月免费"已于 2025-06 下线。
- **其它国内厂商（都不可用，但可以省得你去试）**：京东云函数服务只支持 Node.js 16、仅异步调用、没有 HTTP/cron 触发器、函数不能访问公网，无法承载 PicFast；另外京东云里叫「**云托管**」的产品其实是 IDC 机柜托管、不是容器托管（容器产品是原生容器 / JKE / JCR，其中原生容器要求余额 > 50 元）；青云公有云根本没有 serverless/函数计算产品。两家都不需要绑定银行卡，但也没有可用的长期免费额度。免费 PG 方面，华为、百度、京东、青云**都没有**（百度免费的 GaiaDB 是 MySQL 兼容，与 PicFast 的 PostgreSQL 不兼容）。
- 国内这几家都缺少与"Render 一键按钮"对等的一键部署入口，落地时按 §7 的方式准备迁移（可在 CI 里用直连串执行 `picfast migrate`）。
- 上述额度与限制均为 2026-10 按官方文档核对的结果，多来源于促销页，随时可能变化，动手前请再确认一次。

---

## 5. 各平台部署方式

### 5.1 Fly.io

Fly 的实例是常驻 VM，但支持 `auto_stop_machines`/`auto_start_machines` 在空闲后停机、有请求时自动唤醒（约 2s 级冷启动，`suspend` 模式更快），配合 slim 镜像就是一套"按需计费"的容器方案。

**目录里放一个 `fly.toml`（无需克隆仓库）：**

```toml
app = "picfast"
primary_region = "sin"           # 就近选择区域

[build]
  image = "docker.io/xbeta/picfast:v0.26.0-slim"   # 直接用公开镜像，不做本地构建

[env]
  PICFAST_SERVER_BASE_URL = "https://picfast.fly.dev"
  PICFAST_SERVER_METRICS_ENABLED = "false"
  PICFAST_JOBS_EMBEDDED = "false"
  PICFAST_DATABASE_AUTO_MIGRATE = "false"
  # 想用环境变量在首启建 S3 策略就必须成对给出，且 JSON 必须合法（否则启动失败）；
  # 库里已有策略时这两个变量会被忽略（日志有 WARN），去后台「存储策略」添加即可。
  # PICFAST_STORAGE_DEFAULT_TYPE = "s3"
  # PICFAST_STORAGE_DEFAULT_CONFIG = '{"endpoint":"...","region":"auto","bucket":"...","access_key":"...","secret_key":"...","link_mode":"proxy"}'

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = "stop"    # 空闲停机（缩容到零）
  auto_start_machines = true
  min_machines_running = 0
  [http_service.concurrency]
    type = "requests"
    soft_limit = 200

  [[http_service.checks]]        # 探针走就绪检查；路径必须直接返回 2xx（Fly 不跟随重定向）
    interval = "15s"
    timeout = "5s"
    grace_period = "10s"
    method = "get"
    path = "/readyz"
    [http_service.checks.headers]
      X-Forwarded-Proto = "https"   # 配合 force_https，避免探针被重定向

[deploy]
  release_command = "migrate"    # 覆盖 CMD，镜像 ENTRYPOINT 仍会执行 → picfast migrate

[[vm]]
  size = "shared-cpu-2x"
  memory = "1gb"                 # 图片处理需要内存，见 §2
```

```bash
fly apps create picfast        # 首次创建应用（镜像与配置都来自 fly.toml）
fly secrets set \
  PICFAST_DATABASE_URL='postgres://...' \
  PICFAST_JWT_SECRET='...' \
  PICFAST_STORAGE_DEFAULT_CONFIG='{"endpoint":"...","bucket":"...","access_key":"...","secret_key":"..."}' \
  PICFAST_JOBS_TOKEN='...'
fly deploy
```

- **迁移**：`[deploy] release_command = "migrate"` 会在每次部署前起一台临时 Machine 执行 `/usr/local/bin/picfast migrate`，失败则中止部署。Fly 的文档语义是"覆盖 CMD、不覆盖 ENTRYPOINT"，所以这里直接写子命令即可。临时 Machine 有网络与密钥，但没有卷。
- **定时任务**：`fly machine run xbeta/picfast:v0.26.0-slim --app picfast --schedule hourly jobs run` 只支持 hourly/daily/weekly/monthly。需要**每分钟**投递 webhook 时，用外部调度调用 `POST /internal/jobs/run`（GitHub Actions 最短约 5 分钟，自建 crontab 可以做到每分钟），或参考 Fly 的 Cron Manager 蓝图用一个常驻小应用按计划启动任务 Machine，或把 `PICFAST_JOBS_EMBEDDED` 保持默认（`true`）并让实例常驻。定时 Machine 是独立创建的一次性实例，请用 `-e PICFAST_DATABASE_URL=...` 明确传入它需要的变量（它不挂卷，也不会替你启动应用进程）。
- **注意**：`auto_stop_machines = "stop"` 时后台任务随实例一起停，所以上面把内置调度关掉、改由外部触发更可控。

### 5.2 Railway

**控制台步骤：**

1. New Project → **Empty Project** → Add a Service → **Docker Image**，镜像填 `xbeta/picfast:v0.26.0-slim`。
2. Variables 里填 §3 的通用变量（`PICFAST_SERVER_PORT=8080` 可选，Railway 会注入 `PORT`，PicFast 会自动使用）。
3. Settings → Networking → **Generate Domain**（生成 `*.up.railway.app` 并提供自动 HTTPS），把 Target Port 指向 8080。Railway 在首次生成域名时也会尝试自动探测应用监听的端口。
4. 健康检查填 `/readyz`（成功判定为任意 2xx，默认超时 300s）。注意健康检查请求来自 `healthcheck.railway.app`，应用需要接受该 Host。
5. 迁移：服务设置里的 **Pre-deploy Command** 填 `migrate`（独立容器执行，不挂卷，失败会中止部署；配合 `PICFAST_DATABASE_AUTO_MIGRATE=false`）。它在**独立容器**里运行、不挂载卷，也不依赖应用进程启动；如果日志提示找不到可执行文件，改成完整路径 `/usr/local/bin/picfast migrate`（见 §2 的 distroless 说明）。也可以完全不用钩子，按 §7 在平台外执行迁移。
6. 定时任务：**再建一个服务**（同一镜像），Settings → **Cron Schedule** 填 `*/5 * * * *`，启动命令填 `jobs run`（同样是"覆盖 CMD"的语义；报找不到可执行文件时改成完整路径）。Railway 的 cron 最短间隔是 **5 分钟**，且要求进程执行完立即退出（`picfast jobs run` 正好符合；上一次没结束会跳过本次）。想每分钟跑就用外部调度调用 `POST /internal/jobs/run`。

**注意事项：**

- Railway 的 HTTP 入口有明确限制：请求体必须在 **5 分钟内**上传完成；空闲 5 分钟无数据传输的连接会被关闭；请求头合计上限 32 KB。
- 通过 **80 端口的 POST 会被改写成 GET**，请始终使用 HTTPS 地址。
- 开启 Serverless（缩容到零）后空闲 5–10 分钟入睡，**唤醒后的第一个请求可能返回 502**；对图床这种"随时被外链访问"的场景，建议保持常驻实例（副本数 ≥ 1）或接受偶发一次 502。
- Config as Code（`railway.json` / `railway.toml`）已标记弃用，新部署优先用控制台或 Infrastructure as Code。

### 5.3 Render

**一键部署（Free 档，不需要绑卡）：** 仓库根目录的 [`render.yaml`](../render.yaml) 就是给它用的配置，README 顶部的 Deploy to Render 按钮可直接创建。控制台里对应 New → **Web Service** → Source 选 **Existing Image** → Image URL 填 `docker.io/xbeta/picfast:v0.26.0-slim`（镜像需为 `linux/amd64`，压缩后 ≤ 10 GB）。

```yaml
services:
  - type: web
    name: picfast
    runtime: image
    image:
      url: docker.io/xbeta/picfast:slim
    region: singapore
    plan: free                       # Free 档：0 美元、不需要绑卡
    healthCheckPath: /readyz
    maxShutdownDelaySeconds: 120     # 默认 30s，最大 300s
    envVars:
      - key: PICFAST_DATABASE_URL
        sync: false                  # 创建时粘贴 Neon / Aiven 等免费托管 PG 连接串
      - key: PICFAST_JWT_SECRET
        generateValue: true
      - key: PICFAST_SERVER_METRICS_ENABLED
        value: "false"
      - key: PICFAST_APP_MAX_UPLOAD_BYTES
        value: "10485760"            # 512 MB 内存下先留足图片解码余量
```

- **端口**：Render 注入 `PORT`，web service 默认期望 **10000**；PicFast 会自动跟随 `PORT`，无需配置。服务必须绑定 `0.0.0.0`，Render 只转发一个 HTTP 端口（`PORT` 可用环境变量覆盖，但不要和 `PICFAST_SERVER_PORT` 同时设置成不同值）。
- **健康检查**：默认是 TCP 探测；填了 `healthCheckPath` 后按 HTTP 检查，2xx/3xx 视为健康、5 秒内需响应，连续失败 15 秒摘流、60 秒重启实例。
- **Free 档的取舍**：0.1 CPU / 512 MB 内存；15 分钟无入站流量后休眠、再访问约 1 分钟启动；每月 750 实例小时（休眠时不消耗）。实例休眠期间后台任务不运行。模板已把单文件上传上限降到 10 MiB。
- **数据库**：模板刻意不创建 Render Postgres（免费档创建 30 天后过期），请在创建服务时粘贴 Neon / Aiven 等免费 PG 的连接串；池化串记得加 `default_query_exec_mode=exec`。
- **对象存储**：部署完成后在控制台「存储策略」里添加（Backblaze B2 / Tigris / Neon Object Storage 都不需要绑卡）。
- **迁移**：免费档保持默认的"启动时自动迁移"即可（多实例并发由咨询锁串行化）。付费档可改用 `preDeployCommand: migrate`——它在部署前于独立实例执行一次，注意命令要按 §2 里"覆盖 CMD"的语义写（`migrate`）。
- **需要 cron 或更大内存时**：升级到付费档，并追加一个 cron 服务（Render 的 cron 没有免费档），此时把 Web 服务的 `PICFAST_JOBS_EMBEDDED` 设为 `false`，避免与外部触发重复：

```yaml
  - type: cron
    name: picfast-jobs
    runtime: image
    image:
      url: docker.io/xbeta/picfast:slim
    schedule: "*/5 * * * *"          # 全 UTC
    dockerCommand: jobs run          # 覆盖 Dockerfile 的 CMD → picfast jobs run
    envVars:
      - key: PICFAST_DATABASE_URL
        sync: false
```

- **Docker Command 的语义**：它**覆盖 Dockerfile 的 `CMD`**，参数会追加到镜像 `ENTRYPOINT`，所以 `jobs run` 这样写即可。Render 文档建议"多条命令用 `/bin/sh -c` 串联"——slim 没有 shell，**这类写法在 slim 上不可用**（需要时改用完整镜像）。
- **注意**：镜像型服务**不会**因为 tag 更新自动重新部署（包括 `latest`），升级请手动触发 Deploy latest reference 或重新部署。

### 5.4 Koyeb

```bash
koyeb services create picfast \
  --docker docker.io/xbeta/picfast:v0.26.0-slim \
  --type web \
  --regions fra \
  --instance-type small \
  --ports 8080:http \
  --min-scale 0 --max-scale 1 \
  --env PICFAST_SERVER_BASE_URL=https://picfast-<org>.koyeb.app \
  --env PICFAST_SERVER_METRICS_ENABLED=false \
  --env PICFAST_JOBS_EMBEDDED=false \
  --env PICFAST_DATABASE_AUTO_MIGRATE=false \
  --env PICFAST_DATABASE_URL='{{secret.picfast_db_url}}' \
  --env PICFAST_JWT_SECRET='{{secret.picfast_jwt}}'
```

（参数名以你安装的 `koyeb` CLI 版本为准；控制台里对应"Create Service → Docker"。）

- **端口**：Koyeb 会定义 `PORT`，未显式设置时等于你暴露的最小端口号；端口默认全部关闭，必须显式暴露（`--ports 8080:http`）。
- **健康检查**：默认 TCP 检查；可在端口上配置 HTTP 路径（2xx/3xx 视为健康）。**所有暴露端口的检查都通过后**才会切流量，否则保留上一个健康版本。
- **迁移**：没有发布钩子，用 `koyeb services exec picfast /usr/local/bin/picfast migrate` 在实例内执行（若提示找不到可执行文件，改成 `migrate`，见 §2；也可按 §7 在平台外执行）。
- **定时任务**：Koyeb 文档中没有 cron/定时任务功能，请用外部调度调用 `POST /internal/jobs/run`（GitHub Actions、Cloudflare Worker Cron、自建 crontab 均可），或让 `PICFAST_JOBS_EMBEDDED` 保持默认并保证常驻实例。
- **缩容到零**：需要把最小实例数设为 0；属于预览特性，且**HTTP/2 请求无法唤醒**睡眠中的服务（用 HTTP/1.1 或 HTTPS 正常刷新即可），唤醒约 1–5 秒。
- **卷**：Koyeb 的卷仍是预览特性（仅 `was`/`fra`、1–10 GB、服务副本数必须为 1），不要用于图床数据。

### 5.5 Zeabur

**控制台步骤：**

1. 新建项目 → Add Service → **Docker Images**，Image 填 `xbeta/picfast:v0.26.0-slim`。
2. **Environment Variables** 填 §3 的变量（可用 `${KEY}` 引用同项目的其他变量）。
3. **Ports** 增加一条：Port = 容器监听端口（如 8080），Type = **HTTP**。只有声明过的端口可被访问；HTTP 类型的端口会自动获得带 TLS 证书的域名（`*.zeabur.app`，也可绑自定义域名）。
4. **Settings → Health Check** 填 `/readyz`（默认是 TCP 探测；自定义路径要求 2xx）。
5. **迁移**：在服务 Overview → **Command** 里执行 `/usr/local/bin/picfast migrate`（或 `jobs run` 手动跑一轮后台任务）。
6. **定时任务**：Zeabur 文档没有 cron 功能，需要外部调度调用 `POST /internal/jobs/run`，或让内置调度常驻（保持 `PICFAST_JOBS_EMBEDDED` 默认 true 且服务不挂起）。
7. **缩容**：Zeabur 只支持垂直扩缩容，**没有缩容到零**；不用时可在控制台手动 Suspend 停止计费。

**注意事项：**

- Zeabur 的健康检查默认每 10s 一次、超时 5s、连续失败 3 次判定不健康；健康检查通过前旧版本继续服务（挂载卷时改为先停后起，会有短暂停机）。
- 卷**不能跨服务共享**（重新使用同一 Volume ID 会创建新卷），需要共享数据时官方建议使用对象存储——正好符合 §1.2 的做法。
- 用 Cloudflare 代理时不要选 Full (Strict)，可能因证书问题连接失败。

---

## 6. 客户端 IP 与 `trusted_proxies`

PicFast 依赖 `X-Forwarded-For` / `X-Real-IP` / `X-Forwarded-Proto` 获取真实客户端 IP，用于**按 IP 限流、审计日志、以及登录 Cookie 的 `Secure` 标记**。判定规则（`internal/clientip`）：

1. 只有 TCP 来源地址（容器看到的对端）可信时才会读这些头。**回环、私有网段（10/8、172.16/12、192.168/16、IPv6 ULA）、链路本地地址默认被信任**，无需配置——大多数平台的入口代理就是从这类地址连容器的，所以通常开箱即用。
2. 解析顺序：`CF-Connecting-IP` → `X-Forwarded-For` 中**最右侧的未受信地址** → `X-Real-IP` → 连接地址。`X-Forwarded-For` 是从右往左找**第一个不在受信列表里**的地址，因此列表里只应放**代理跳数**。

需要显式配置时（来源是公网地址）用逗号分隔的 CIDR：

```bash
PICFAST_SERVER_TRUSTED_PROXIES="10.0.0.0/8,172.16.0.0/12,fc00::/7"
```

> ⚠️ **不要写 `0.0.0.0/0`**：它会把所有地址都当成"受信的代理跳数"，于是 `X-Forwarded-For` 里再也找不到可用的客户端地址，最终只能回退到 `X-Real-IP` 或代理地址——对按 IP 限流没有任何帮助。正确做法是只加入平台代理实际使用的网段。

**各平台现状**（撰写时按官方文档核对）：

| 平台 | 客户端 IP 情况 |
|------|----------------|
| Fly.io | Fly 另外提供 `Fly-Client-IP`（PicFast 目前不读取），并在 `X-Forwarded-For` **末尾附加你自己应用的公网 IP**。由于右端优先，PicFast 会取到该公网 IP。修法：把 `fly ips list` 里的应用公网 IP 加进 `PICFAST_SERVER_TRUSTED_PROXIES`，右端就会被跳过、露出真实客户端 IP。 |
| Railway | 文档明确提供 `X-Real-IP` 作为客户端 IP，且 `X-Forwarded-Proto` 恒为 `https`；`X-Forwarded-For` 不在其文档的头列表里，PicFast 会在 XFF 无可用地址时回退到 `X-Real-IP`。 |
| Render / Koyeb / Zeabur | 未公开转发头行为，也没有代理网段。请按下面的方法验证后再决定是否配置。 |

**验证方法**：登录后台看审计日志里的来源 IP，或对某个接口连续请求触发限流（限流器按实例计数），看来源 IP 是不是"所有访客共用一个地址"——是则说明取到的是代理地址。若确实取不到真实 IP，可按平台文档给出的代理网段补进 `trusted_proxies`；同时确认站点地址是 `https://`，否则登录 Cookie 不会带 `Secure` 标记（`Secure` 由受信的 `X-Forwarded-Proto: https` 或直连 TLS 决定）。

---

## 7. 数据库迁移怎么跑

迁移已嵌入二进制，三种方式任选：

1. **启动时自动迁移（默认）**：每个实例启动时都会尝试迁移。多实例并发是安全的——golang-migrate 用 PostgreSQL 会话级咨询锁串行执行，只有一个实例真正应用迁移，其余等待后发现无变更即继续。缺点是每次冷启动都会多花一点时间并占用一次数据库连接。
2. **平台发布钩子**：Fly 的 `[deploy] release_command`、Railway 的 Pre-deploy Command、Render 的 `preDeployCommand` 都能在部署前执行一次 `picfast migrate`（记得同时设 `PICFAST_DATABASE_AUTO_MIGRATE=false`）。Fly 是"覆盖 CMD、保留 ENTRYPOINT"的语义，直接填 `migrate`；其余平台若提示找不到可执行文件，改成完整路径 `/usr/local/bin/picfast migrate`（见 §2）。slim 镜像无 shell，不要写 `sh -c "picfast migrate"` 或 `a && b` 这类形式。Koyeb / Zeabur 没有发布钩子，可在实例内执行：`koyeb services exec picfast migrate`、Zeabur 控制台 Command 里执行 `/usr/local/bin/picfast migrate`（Command 界面若通过 shell 执行，请改用完整镜像）。
3. **本地/CI 执行**：不依赖平台能力，用直连连接串在任意机器上跑同一镜像：

```bash
PICFAST_DATABASE_URL='postgres://user:pass@direct-host:5432/picfast?sslmode=require' \
  docker run --rm -e PICFAST_DATABASE_URL xbeta/picfast:v0.26.0-slim migrate
```

> **池化连接串不能跑迁移**：迁移锁是会话级的，事务模式连接池（Supabase `:6543`、Neon `-pooler`）无法可靠持有它。迁移请用直连/会话模式串；应用本身可以用池化串（记得加 `default_query_exec_mode=exec`）。

若应用账号没有建表/改表权限（部分托管 PG 分配受限账号），请保持 `PICFAST_DATABASE_AUTO_MIGRATE=false`，用有 DDL 权限的账号单独执行 `picfast migrate`。

---

## 8. 上线验收清单

```bash
# 1) 探针
curl -fsS https://pics.example.com/healthz     # 期望 200 {"status":"ok"}
curl -fsS https://pics.example.com/readyz      # 期望 200（数据库不通则为 503）

# 2) 上传一张图片，确认对象存储与缩略图
#    上传成功后：图片 URL 应指向你的存储/CDN 域名；/t/{md5}.png 能返回缩略图
curl -fsS -X POST https://pics.example.com/api/v1/upload \
  -H "Authorization: Bearer <API Token>" -F "file=@test.jpg"

# 3) 触发一轮后台任务（外部定时器用的是同一路径）
curl -fsS -X POST https://pics.example.com/internal/jobs/run \
  -H "Authorization: Bearer $PICFAST_JOBS_TOKEN"
```

另外确认：

- [ ] 实例重建/重新部署后，**已上传的图片仍然可访问**（说明没有落在本地磁盘上）
- [ ] 后台能看到存储策略指向 S3/R2，而不是默认的本地存储
- [ ] 已配置外部定时触发（或在实例常驻时可接受内置调度）
- [ ] 站点地址为 `https://`，登录后 Cookie 带 `Secure`
- [ ] 若开启了缩容到零，冷启动后的首次上传/登录可用

---

## 9. 常见问题

| 现象 | 原因与处理 |
|------|------------|
| 重新部署后上传的图片全部 404 | 用的是默认 `local` 存储，落在临时磁盘上。请配置 S3/R2（§1.2）后重新上传；旧文件需要在实例被回收前自行导出（PicFast 没有"在存储策略之间搬迁对象"的命令，可用 `picfast maintenance doctor` 巡检缺失对象）。 |
| 实例启动即退出 | 连不上数据库。检查 `PICFAST_DATABASE_URL`、`sslmode`、平台是否放通出网（5432），以及托管 PG 是否处于挂起状态。 |
| 冷启动后第一次请求失败 | 数据库同时被唤醒或实例刚起。探针用 `/readyz`、平台就绪判定放宽一些，或设置最小实例数 ≥ 1。 |
| 上传大图时实例 OOM | slim 用标准库解码，峰值内存约为像素数的 3 倍（见 §2）。提高实例内存到 ≥1 GB，或关闭图片处理 / 下调 `PICFAST_APP_MAX_UPLOAD_BYTES`。 |
| 上传被平台网关拒绝（413/超时） | 多数平台没有公开请求体上限；PicFast 默认允许 50 MiB。请下调 `PICFAST_APP_MAX_UPLOAD_BYTES`，并确认平台超时 ≥ 60s。 |
| webhook 长时间不投递 | 内置调度被关（`PICFAST_JOBS_EMBEDDED=false`）但没有配置外部触发，或触发间隔太长。投递延迟等于触发间隔，建议每分钟一次。 |
| 限流比配置更严 / 日志里 IP 都是同一个 | 取到的是平台代理地址，见 §6。 |
| `.webp` 的 `@w_`/`@q_` 变体返回原图 | slim 不支持对 WebP 源图动态缩放/压缩（§2）。需要时改用完整镜像。 |
| 迁移报 "dirty" 或连接被中断 | 用了事务模式池化连接串。迁移改用直连/会话模式串（§7）。 |
| 升级后仍是旧版本 | 这些平台不会因为镜像 tag 移动而自动重新部署（Render、Koyeb 都是如此）。手动触发一次部署，或改用固定版本标签 + 明确的重部署流程。 |
| 一键部署时被要求绑银行卡 | 旧模板里含 cron 服务与付费档（Render 的 cron 没有免费档）。现在 `render.yaml` 只用 Free 档、不建 cron 与 Postgres，无需绑卡。 |
| Free 档（512 MB）上传稍大的图就失败 | 纯 Go 解码峰值约为 `3 × 宽 × 高 × 4` 字节，模板已把上传上限降到 10 MiB。继续下调 `PICFAST_APP_MAX_UPLOAD_BYTES`、关闭图片处理，或升级到内存 ≥ 1 GB 的档位。 |
| 免费 PG 的月度额度很快用完 | 内置后台任务每 5 秒轮询数据库，会让 Neon 这类空闲缩容的库无法休眠。设 `PICFAST_JOBS_EMBEDDED=false` 改用外部定时触发，数据库即可在空闲 5 分钟后缩容到零。 |
| 实例休眠一段时间后 webhook 没投递 | Free 档 15 分钟无入站流量会休眠，实例内调度随之暂停。需要准点投递请升级付费档并加 cron 服务，或用外部调度打 `POST /internal/jobs/run`。 |
