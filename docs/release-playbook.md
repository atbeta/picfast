# Release Playbook

本文件描述 PicFast 的镜像发布策略：本地快速开发镜像（dev）与 CI 正式发布镜像（release）分离执行。

## 1) 策略总览

- `dev`：允许本地构建并推送，用于临时验证与联调。
- `release`：必须通过 CI（`release-please` + Docker 发布工作流）执行，用于正式版本。
- `latest`：仅用于正式发布链路，不用于临时开发镜像。

## 2) 本地发布 dev 镜像

推荐标签：

- 不可变标签：`dev-<yyyymmdd-HHMM>-<shortsha>`
- 可变标签（可选）：`dev-main`

示例：

```bash
TAG="dev-$(date +%Y%m%d-%H%M)-$(git rev-parse --short HEAD)"
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t xbeta/picfast:${TAG} \
  -t xbeta/picfast:dev-main \
  --push .
echo "${TAG}"
```

## 3) 服务器验证 dev 镜像

将部署环境中的 `PICFAST_IMAGE` 指向新标签后重启：

```bash
docker compose pull
docker compose up -d
docker compose ps
```

## 4) 正式发布 release 镜像

正式发布流程：

1. 合并业务提交到 `main`
2. 由 `release-please` 生成并提交 `chore(main): release x.y.z`
3. 推送 `vX.Y.Z` 标签触发 Docker 发布
4. CI 自动发布 `latest` 与 semver 标签

## 5) 回滚

回滚优先使用上一正式版本镜像：

```bash
# 例如回滚到 v0.2.0 对应镜像标签
# 在 .env 中设置 PICFAST_IMAGE=xbeta/picfast:0.2.0 后执行
docker compose pull
docker compose up -d
```

## 6) 清理建议

- 保留最近 N 个 dev 标签（例如 20 个），定期清理历史开发镜像。
- 正式版本标签与 `latest` 长期保留。
