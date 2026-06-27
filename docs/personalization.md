# PicFast 个性化体系

> 状态：当前实现（2026-06）
> 关联：`docs/theme-system.md`（主题系统基线 + 维护指南）、`docs/theme-migration.md`（v0.16 主题升级指南）

## 1. 体系概览

经过几轮简化，PicFast 的"个性化"被收敛成 3 个 type × 2 个 layer 的清晰结构：

| Type    | 归属 layer   | 现状                                                         |
| ------- | ------------ | ------------------------------------------------------------ |
| `theme` | admin only   | 自定义 CSS 一项，注入站点 `<style>` 标签                    |
| `mode`  | user（本地） | 亮/暗/系统切换、界面语言；只存 localStorage，不进 user.settings |
| `workflow` | user only   | 上传默认值（策略 / 相册 / 权限）、上传时图片处理          |

外加一个不属于"个性化"、但通常和"账户"放在一起的：

| Type      | 归属 layer | 现状                                                       |
| --------- | ---------- | ---------------------------------------------------------- |
| `account` | user only  | 资料、改密码、关联账号、存储用量（在 `/console/account`） |

之前的 5 个 type 里被砍掉的：

- `output`（admin 默认 + user 覆盖的复制格式 / 模板）：commit `af67068` 整 type 移除。结果是上传结果卡片只复制 Markdown，其他格式用户当场选。
- `theme_override`（用户级主题覆盖）：commit `278b77a` 移除。用户的偏好只剩 `mode`（亮暗 + 语言）和 `workflow`。

## 2. Layer × Type 矩阵（最终态）

```
                │ site (admin)                              │ user (account)
────────────────┼───────────────────────────────────────────┼─────────────────────────────────
theme           │ theme_config.custom_css 注入站点 <style> │ —
                │                                           │
mode            │ theme_config.mode（未登录访客的兜底）       │ light / dark / system（localStorage）
                │                                           │ language（i18n localStorage）
                │                                           │
workflow        │ allow_user_image_processing ⭐            │ default_strategy / default_album
                │ skip_image_processing ⭐                  │ default_permission
                │                                           │ image_processing
                │                                           │ （受 allow_user_image_processing 控制）
────────────────┼───────────────────────────────────────────┼─────────────────────────────────
account         │ (admin 在 admin 页面管理其他用户)          │ name / password / email_verified
                │                                           │ OAuth identities
                │                                           │ 存储用量（只读展示）
```

⭐ = lock flag：决定 `workflow` 行是否对 user 暴露（仅 admin 可改）。

## 3. 单一主题（admin theme）

### 3.1 模型

- `site_settings.theme_config` JSONB 字段，结构：
  ```json
  { "mode": "system", "custom_css": "" }
  ```
- `mode`：可选 `light` / `dark` / `system`，作为未登录访客的兜底；登录后由 `next-themes` 用户本地选择接管。
- `custom_css`：原样注入到站点的 `<style id="picfast-site-theme">` 标签，在内置主题的变量之后生效。

### 3.2 6 个预设的历史

之前存在 `moe / cyber / pixel / terminal / fresh / default` 6 套主题预设、token 调色板、JSON 主题包导入导出、用户级 `theme_override`。这一套被 commit `278b77a` 整体砍掉，理由是：

- 实际部署中很少有人换预设，6 套设计在大多数实例里没被用过。
- 想要细节差异化的用户，custom_css 已经足够。
- 砍掉后 admin 表单只留 1 个 textarea，极大降低配置门槛。

迁移路径见 `docs/theme-migration.md`。

## 4. 用户级 mode（亮暗 + 语言）

| 项       | 存储位置                | 说明                                       |
| -------- | ----------------------- | ------------------------------------------ |
| 亮暗模式 | `localStorage.theme`    | `next-themes` 维护，未登录也用             |
| 语言     | `localStorage.i18nextLng` | i18next 维护                               |

- `site.theme_config.mode` 只影响**未登录访客**的兜底。
- 登录后 user 本地切换的 mode 优先；不会写回 server。
- admin 想强制全站亮暗只能通过 custom_css。

## 5. 用户级 workflow

### 5.1 上传默认值

| 字段              | DB 位置                              | 默认值            |
| ----------------- | ------------------------------------ | ----------------- |
| `default_strategy` | `users.settings.default_strategy`   | 0（跟随分组） |
| `default_album`    | `users.settings.default_album`      | 0（个人相册）   |
| `default_permission` | `users.settings.default_permission` | 1（公开）       |

旧 localStorage 中的 `default_strategy_id / default_album_id / default_permission` 会在 `usePersonalization()` 启动时做一次性的 best-effort 迁移（`migrateLocalStorageToServer`，见 `web/src/lib/personalization.ts`），失败时回退到 localStorage 重试。

### 5.2 上传时图片处理

`users.settings.image_processing` 子对象：

```jsonc
{
  "image_save_quality": 85,            // 1-100，JPEG / WebP 编码质量
  "image_save_format": "origin",        // origin / jpeg / png / webp
  "is_strip_exif": true,                 // 剥离 EXIF
  "is_enable_watermark": false,          // 是否加水印
  "watermark_configs": {                // is_enable_watermark=true 时生效
    "text": "",
    "position": "bottom-right",          // 5 个角
    "font_size": 28,
    "color": "#FFFFFF",
    "opacity": 0.6
  }
}
```

UI 入口在上传页右上角齿轮弹窗。`web/src/components/image-processing-dialog.tsx`。

### 5.3 管理员 lock flag

`site_settings.allow_user_image_processing` 和 `skip_image_processing` 控制：

- `skip_image_processing = true` → 全站跳过处理流水线（最高优先级，覆盖一切）
- `skip = false` 且 `allow_user = false` → 用户弹窗字段全部 disabled，使用系统默认（quality 85 / origin / strip-exif on / no watermark）
- `skip = false` 且 `allow_user = true` → 用户可自由配置（默认行为）

## 6. account

`/console/account` 是 user only 的"身份 + 凭据"页。3 个 section：

- 存储用量（只读）
- 个人资料（昵称 + 改密码）
- 关联账号（条件显示：仅当至少一个 OAuth provider 被启用）

不在个性化体系里，但和"账户设置"语义相邻。

## 7. 设计纪律

经过几次 commit 收敛，几个原则被验证：

1. **每个 preference 必须有且仅有一个 owner**。Admin default + user override 这种模式（曾经的 `output`、`theme_override`）被证明维护成本 > 价值。
2. **不要给 user 调色板**。`theme_override` 时代的 5 套调色板/密度/动效选择几乎没人动，反而增加了表面积。
3. **复杂功能 > 配置功能**。上传时处理保留（user 实际会用），复制格式预设砍掉（user 实际不会设默认值）。
4. **每个 type 单一 UI 入口**。`theme` = admin site settings；`mode` = header 切换器；`workflow` = 上传页弹窗；`account` = 独立页。`/console/settings` 这个"什么都有"的页面被拆分。
5. **JSONB 加键不删键**。所有这轮重构都没有破坏存量数据（数据迁移是数据清理，不是 schema 变更）。
