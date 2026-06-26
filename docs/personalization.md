# PicFast 个性化能力优化设计

> 状态：设计稿（待评审）
> 范围：前端 + 后端增量，向后兼容，无破坏性 migration
> 关联：`docs/theme-system.md`（现有主题系统基线）

## 1. 背景与现状

PicFast 当前有 6 套主题预设、token 调色板、自定义 CSS、JSON 主题包导入/导出（详见 `docs/theme-system.md`），但实际使用中"个性化能力"被切成了 5 块互不相通的东西：

| 模块 | 范围 | 路径 |
|---|---|---|
| 站点主题（admin） | 全站视觉 + 公共面 | `web/src/lib/theme-config.ts` + `pages/console/admin/settings/appearance-settings-page.tsx` + `components/site-theme-runtime.tsx` |
| 亮/暗切换 | 所有用户 | `components/theme-switcher.tsx`（仅 light/dark 二态） |
| 用户上传处理 | 单用户 | `pages/console/settings-page.tsx` 内嵌 |
| 复制偏好 | 单用户（覆盖站点） | `lib/use-copy-preferences.ts` |
| 上传默认值 | 单用户 | `pages/console/upload-page.tsx` 直接读 `localStorage` |
| 语言 | 单用户 | `components/language-switcher.tsx` + `i18n/index.ts` |

### 1.1 痛点

1. **token 覆盖不完整**：`ThemeTokenSet` 只声明 16 个 token，但 `index.css` 实际定义更多（`--popover-*`、`--warning/*`、`--success/*`、`--info/*`、`--sidebar-primary*` 等）。换预设时这些字段不会被重涂，视觉上"换了一半"。**解决方向：v0 先做 grep 审计（`:root` 定义 vs 组件引用），基于差集扩 token，不做理论清单补全。**
2. **user 没有主题选择权**：theme-switcher 只是二态；站点主题是 admin-only，普通用户只能继承管理员的 preset。
3. **视觉轴分散且靠 magic class 维系**：`upload_style / card_style / button_style / density / motion` 语义在 `theme-config.ts`，实现靠 `index.css` 里 `html[data-pf-*]` 选择器去匹配硬编码在组件里的 magic class（`pf-upload-zone` / `pf-result-card` / `pf-auth-card` / `pf-primary-button` / `pf-site-logo` / `pf-public-glow`）。前 3 项是组件级枚举样式，本质不属于站点主题；后 2 项的全局生效问题应通过 CSS 变量解决。**解决方向：v2 删除 `upload/card/button_style` 3 项；density/motion 改用 `:root` CSS 变量 + `calc()` 驱动，全局零 opt-in。**
4. **用户偏好分裂在 localStorage 和服务器两边**：上传默认值在 `localStorage`（换设备就丢），复制偏好走 `user.settings`（server），语言和亮暗各自走独立的 `localStorage`。无统一概念。
5. **settings 页是大杂烩**：`settings-page.tsx` 565 行堆了存储用量、资料、密码、默认策略、图像处理（含水印 6 个字段）、复制偏好、OAuth 账号，没有分组/折叠。
6. **density / motion 实际无感**：`index.css` 只对 `.pf-upload-zone` 应用 padding 和 transition，对表格、卡片、列表行完全没生效。**解决方向：v2 改为 CSS 变量驱动，组件用 `calc(0.75rem * var(--pf-density))` 和 `transition-duration: var(--pf-motion-duration)`，零组件 opt-in 即全局生效。**

## 2. 目标与非目标

### 2.1 目标

- 引入**统一的"个性化"概念**，前端单一 hook 出口，后端单一 JSONB 出口。
- 明确**配置所有权（layer）**和**配置类别（type）**两个维度，矩阵化所有 key。
- 把分散在 5 个 localStorage + 2 个 server 字段的个人偏好**全部收敛**。
- 让普通用户**能选择自己的主题预设、亮暗、密度、动效**。
- 主题预设真正覆盖整个 UI（不是只换一半）。
- 整套改造**不引入破坏性 migration**，全部增量、可分阶段回滚。

### 2.2 非目标

- 多主题同屏（A/B 主题）——单实例单主题先做扎实。
- 完整 CSS 沙箱 / 用户自定义选择器——`custom_css` 保留给管理员。
- 自动跟随系统壁纸 / Material You 风格的实时调色。
- 用户级调色板（普通用户不能改 token，避免无意义乱涂）。
- 重写 `site_settings` 表结构。

## 3. 顶层架构：Layer × Type 二维矩阵

### 3.1 两个 Layer（配置所有权）

| Layer | 谁配 | 存储位置 | 作用 |
|---|---|---|---|
| `site` | 管理员 | `site_settings`（JSONB 字段 + 列） | 整站的基线、品牌、策略 |
| `user` | 当前账号 | `users.settings`（JSONB） | 个人覆盖项 |

**运行时合并规则（贯穿全局）**：

```
effective = { ...siteDefaults, ...userOverrides(只对允许 override 的 key) }
```

每条 key 必须显式回答 3 个问题：

1. 它属于哪个 layer？（`site` / `user` / 两者）
2. 用户能不能 override？（`overridable: true | false | conditional`）
3. 当用户没设置时，回退到哪？（`site` 默认 / 硬编码默认 / 拒绝）

### 3.2 五个 Type（配置类别）

| Type | 含义 | 典型问题 |
|---|---|---|
| `theme` | 视觉外观：色板、字体、半径、动效、logo 形状、背景 | "页面长什么样" |
| `mode` | 交互模式：亮/暗、系统跟随、密度、动效档位、语言 | "页面怎么用" |
| `output` | 输出格式：复制格式、复制模板 | "分享出去长什么样" |
| `workflow` | 上传/处理工作流：默认策略、相册、权限、图像处理、水印 | "上传时怎么走" |
| `account` | 身份与凭据：昵称、密码、OAuth、邮箱 | "我" |

> `account` 不是"个性化"（无视觉/行为外延），但和"个人偏好"共享同一份"个人设置"页面，需在同一体系内管理。

### 3.3 完整矩阵

```
                │ site (admin)                                       │ user (account)
────────────────┼────────────────────────────────────────────────────┼─────────────────────────────────────────────
theme           │ app_name, favicon, logo_shape, logo asset,         │ theme preset override
                │ public.background_image, public.background_style,  │ (palette/token 不开放给普通用户：
                │ theme preset + tokens + custom_css                 │   安全 + 视觉一致性考量)
────────────────┼────────────────────────────────────────────────────┼─────────────────────────────────────────────
mode            │ theme_config.mode (未登录访客的兜底)                │ light / dark / system, language
                │ public.density / motion (CSS 变量站点兜底)          │ density, motion (CSS 变量个人覆盖)
────────────────┼────────────────────────────────────────────────────┼─────────────────────────────────────────────
output          │ default_copy_format, copy_template (站点级 fallback)│ default_copy_format (override),
                │                                                    │ copy_template (override)
────────────────┼────────────────────────────────────────────────────┼─────────────────────────────────────────────
workflow        │ default_image_ttl, guest_image_ttl,                │ default_strategy, default_album,
                │ allow_guest_upload, guest_capacity_bytes,         │ default_permission, image_processing
                │ allow_user_image_processing ⭐, skip_image_processing ⭐│ (仅当 allow_user_image_processing=true)
────────────────┼────────────────────────────────────────────────────┼─────────────────────────────────────────────
account         │ (admin 在 admin 页面管理其他用户)                   │ name, password, email_verified, OAuth
```

⭐ 为 **lock flag**——决定 workflow 行是否对用户暴露。

> **不在矩阵中的项**（v2 删除）：`public.upload_style / card_style / button_style`——本质是组件级枚举样式，由组件自身决定，admin 若需特殊样式走 `custom_css`。site theme 不再承载。

### 3.4 Override Boundary（边界控制）

| 状态 | 含义 | 例子 |
|---|---|---|
| `overridable: true` | 站点设默认，用户可改 | copy_format, copy_template, theme preset override, density, motion |
| `overridable: false` | 站点锁死，用户不能动 | allow_user_image_processing, skip_image_processing, allow_guest_upload |
| `overridable: conditional` | 站点 flag 决定是否暴露 | image_processing → 仅当 `allow_user_image_processing=true` 时才在用户设置页出现 |

实现机制：站点 lock flag（`allow_user_image_processing` / `skip_image_processing`）由前端 `usePersonalization()` 在构造时读取，**直接决定用户面板要显示哪些字段**。这避免了"用户在 UI 改了一通，保存时却被服务端拒绝"。

**`density / motion` 特殊说明**：两轴既是 override key，也是 CSS 变量驱动的运行时值。`usePersonalization()` 把合并结果（site default ?? user override ?? 硬编码 default）解析为：

```css
:root {
  --pf-density: 1;             /* compact=0.75, comfortable=1, spacious=1.25 */
  --pf-motion-duration: 150ms; /* none=0, subtle=150ms, playful=300ms */
}
```

组件用 `calc(0.75rem * var(--pf-density))` / `transition-duration: var(--pf-motion-duration)` 自动响应，**零 opt-in**。`SiteThemeRuntime` 唯一职责就是把这几个变量写进 `:root`。

## 4. 数据模型

### 4.1 `site_settings`（已经在用，不动 schema）

现有字段全部保留，**不破坏**：

| 字段 | 归属 type | override |
|---|---|---|
| `app_name`, `favicon_url` | theme | admin only |
| `theme_config JSONB` | theme + mode | admin only |
| `default_copy_format`, `copy_template` | output | overridable |
| `allow_guest_upload`, `guest_capacity_bytes`, `guest_image_ttl` | workflow | admin only |
| `allow_user_image_processing`, `skip_image_processing` | workflow（lock flag） | admin only |
| `default_image_ttl` | workflow | admin only |
| `footer_*` | theme | admin only |
| `analytics_*` | （运营） | admin only |

可选追加（不在本设计强制）：`personalization_schema_version INT DEFAULT 1`（未来大改用）。

### 4.2 `users.settings` JSONB 内部 schema

引入"分类型"内部命名。**加 key，不删 key**，完全向后兼容：

```jsonc
{
  // account（已有，未动）
  "default_strategy": 3,
  "default_album": 12,
  "default_permission": 1,
  "image_processing": { ... },

  // output（已有 key）
  "default_copy_format": "markdown",
  "copy_template": "![{name}]({url})",

  // theme（新增，v1 阶段）
  "theme_override": {
    "preset": "moe",                 // 可选；不设 = 跟随 site
    "mode": "dark",                  // light/dark/system
    "density": "comfortable",        // compact/comfortable/spacious
    "motion": "subtle"               // none/subtle/playful
  },

  // mode（新增，v1 阶段）
  "language": "zh-CN"                // 可选；前端 i18n localStorage 仍作 fallback
}
```

> `UpdateProfile` 是 settings 整体覆盖（`internal/handler/user_handler.go:101-103`），前端在 `usePersonalization` 内统一做 merge，避免丢键。

### 4.3 客户端 fallback 链

`usePersonalization()` 解析顺序（每个 key 一致）：

```
1. user.settings[key]           ← 最高优先级（如果存在且未被 lock flag 屏蔽）
2. site_settings[key]           ← 站点默认
3. 硬编码 default                ← 兜底
```

`localStorage` 仅作为"未登录访客的临时偏好"（mode/language），登录态一律从服务端读。

### 4.4 显式 NOT in this system

下面这些**不进入个性化体系**，避免被误归类：

| 项 | 归属 | 原因 |
|---|---|---|
| `groups.configs`（rate limit、文件类型、组配额） | policy | 是策略/配额，不是偏好 |
| 存储后端配置（S3/Kodo/OSS 的密钥、bucket） | infra | 管理员的运维配置 |
| 图像处理底层算法选项 | policy | 暴露 quality/format 给用户足够 |
| 审核模式、审计日志 | governance | 治理层 |
| 站点元数据（app_name、SEO 描述） | brand | 是品牌/对外文案，不属于个性化 |

## 5. 主题包范围

主题包是 admin 共享/备份用（v0 已有），以及 v3 的 user 共享用。

### 5.1 站点主题包（admin，`scope: "site"`）

进包：

| 字段 | 来源 layer × type |
|---|---|
| `preset` | site × theme |
| `mode` | site × mode（给未登录访客用） |
| `tokens.light.*` | site × theme |
| `tokens.dark.*` | site × theme |
| `public.background_image` | site × theme |
| `public.background_style` | site × theme |
| `public.logo_shape` | site × theme |
| `public.density` | site × mode（CSS 变量站点兜底） |
| `public.motion` | site × mode（CSS 变量站点兜底） |
| `custom_css` | site × theme（管理员特权） |

不进包：

- `language` —— user × mode
- `default_copy_format` / `copy_template` —— user × output
- `default_strategy` / `default_album` / `default_permission` —— user × workflow
- `image_processing` —— user × workflow
- ~~`public.upload_style / card_style / button_style`~~ —— v2 删除，不属于站点主题

### 5.2 用户主题包（v3，`scope: "user"`）

只装 4 个轴：

```jsonc
{
  "scope": "user",
  "preset": "moe",           // 不写 = 跟随站点
  "mode": "dark",            // light/dark/system
  "density": "comfortable",  // compact/comfortable/spacious
  "motion": "subtle"         // none/subtle/playful
}
```

**显式不带**：tokens、custom_css、background_image、logo_shape、background_style。`upload_style / card_style / button_style` 已从 site 主题包中删除（v2），user 包自然也不再带。

解析器侧：`parseThemePackage(raw)` 返回 `{ config, scope: 'site' | 'user' }`；`scope: "user"` 时**只接受** 4 个键，其他键视为错误（不要"宽容忽略"，避免用户误以为导入成功但站点字段被吞了）。

### 5.3 解析器扩展

`web/src/lib/theme-package.ts`：

- `exportThemePackage(config, scope: 'site' | 'user')`
- `parseThemePackage(raw): { config, scope }`
- `THEME_PACKAGE_KEYS` 按 scope 分两套常量
- `ThemePackageError` 报错信息明确指出"该字段在 user scope 不允许"

后端在 v3 阶段需要 `internal/handler/user_handler.go` 加 `validateUserSettings` 对 `theme_override` 形状的校验（与 `validateThemeConfig` 对称）。

## 6. 实施计划

按"先收敛、再扩展"分 3 个独立 PR（外加 v3 可选），每 PR 都向后兼容。

### 6.1 v0：单一事实源 + 审计驱动的 token 覆盖

**目标**：把分散在 5 个地方的东西收敛到一个 hook；基于**组件实际使用情况**扩 `ThemeTokenSet`，不按理论清单补全。

**第一步（必做，提交前完成）**：token 差距审计

```bash
# :root / .dark 实际定义的所有 CSS 变量
grep -oE -- '--[a-z][a-z0-9-]*' web/src/index.css | sort -u
# 组件中实际引用的 CSS 变量（排除 index.css 自己）
rg -- '--[a-z][a-z0-9-]*' web/src -t tsx -t ts \
  | grep -oE -- '--[a-z][a-z0-9-]*' | sort -u
```

两者的差集 = 已定义但无消费者的 token（如 `--chart-1..5` 若无 chart 组件则不进 `ThemeTokenSet`）。审计结果写进 PR 描述。

| 文件 | 改动 |
|---|---|
| `web/src/lib/personalization.ts` (new) | 定义 `PersonalizationState` 类型 + `getEffectiveTheme(site, userOverride)` + 5 个 type 的常量表（每个 key 标 layer / overridable / storage） |
| `web/src/lib/use-personalization.ts` (new) | `usePersonalization()` hook：合并 `site-config` query + `auth-context` 用户的 `theme_override`，产出 5 个分 type 切片 |
| `web/src/lib/theme-config.ts` | `ThemeTokenSet` 扩到**审计结果**需要补的项（不预设 24 项），6 个预设按差集补值 |
| `web/src/index.css` | `:root` / `.dark` 补齐新进 `ThemeTokenSet` 的 token |
| `web/src/components/site-theme-runtime.tsx` | 内部改用 `usePersonalization()`，仍写 `<style>` 和 dataset，**对外行为不变** |
| `web/src/lib/use-copy-preferences.ts` | 改为 `usePersonalization().output` |
| `web/src/lib/theme.tsx` | 保留 `useTheme` 包装，内部透传到 `usePersonalization().mode` |
| `web/src/pages/console/upload-page.tsx` | localStorage 读取改为 `usePersonalization().workflow`（v0 只读不写，保留 localStorage 写回作 fallback） |

**验证**：
- `make test && make lint`
- 6 个预设切换，审计纳入的 token 都生效、dark mode 仍正常
- 现有 `admin/settings/appearance-settings-page` 不需要改

**提交**（拆 2 个）：
1. `refactor(web): introduce personalization module as single source of truth`
2. `feat(theme): extend token set based on consumer audit`

### 6.2 v1：用户主题覆盖 + 上传默认值搬服务端

**目标**：让普通用户能选自己的 preset/mode/density/motion，并把上传默认值从 localStorage 迁到服务器。

后端（最小侵入，但**全字段覆盖**）：

| 文件 | 改动 |
|---|---|
| `internal/handler/user_handler.go` | 新增 `validateUserSettings(raw)` 函数，**按 `domain.UserSettings` struct 全量校验**（6 字段：`default_strategy / default_album / default_permission / image_processing / default_copy_format / copy_template` + v1 新增的 `theme_override / language`），同时拒绝未知 key。一次性消除 schema/code 漂移。 |
| `internal/service/upload_service.go` | `uploadUserSettings` 扩字段读 `default_album` / `default_permission`（v0 已声明但没消费） |

前端：

| 文件 | 改动 |
|---|---|
| `web/src/lib/personalization.ts` | 加 `ThemeOverride` 类型 + `saveThemeOverride()` + 整体 merge helper；新增 `migrateLocalStorageToServer()` |
| `web/src/components/theme-switcher.tsx` | 升级为 popover：3 段（mode 三态 / preset 缩略图 / density+motion） |
| `web/src/pages/console/upload-page.tsx` | 上传默认值从 `usePersonalization().workflow` 读；首次挂载调用 `migrateLocalStorageToServer()` |
| `web/src/lib/i18n/index.ts` | `language` 改读 `usePersonalization().mode.language`，写回 server；保留 localStorage 兜底（未登录） |

**localStorage 迁移（带失败恢复）**：

```ts
// 读取优先级
function readPref(key) {
  return server.value ?? localStorage.get(key) ?? hardcodedDefault
}

// 迁移触发（首次挂载）
async function migrateLocalStorageToServer() {
  if (localStorage.get('pf-migrated-v1') === '1') return  // 已迁移过
  const local = readAllLocalPrefs()
  if (Object.keys(local).length === 0) {
    localStorage.set('pf-migrated-v1', '1')
    return
  }
  try {
    await updateProfile({ settings: mergeWithServer(server.settings, local) })
    Object.keys(local).forEach(k => localStorage.removeItem(k))
    localStorage.set('pf-migrated-v1', '1')
  } catch {
    // 保留 localStorage，下一次挂载重试；当前 session 走双读 fallback
    toast.warning('偏好同步失败，将在下次登录重试')
  }
}
```

关键点：
- **双读**（server ?? localStorage ?? default），迁移失败期间用户不丢偏好
- **失败保留**（不删 localStorage），下次挂载重试
- **迁移标志**（`pf-migrated-v1`），避免每次挂载都重试 + 显式记录 schema 版本

**验证**：
- 6 个预设的用户切换；切换后刷新仍生效
- 退出登录 → 仍按站点 mode 走
- 旧 localStorage 值的设备首次进入能自动迁移；模拟 POST 失败时偏好不丢
- 已有的复制偏好、图像处理设置都还在
- `validateUserSettings` 拒绝未知 key / 错误类型（补 Go 单测）

**提交**（拆 2 个）：
1. `feat(user): persist theme override and language in user profile`
2. `refactor(upload): move default strategy/album/permission from localStorage to user settings`

### 6.3 v2：删除 magic class + CSS 变量驱动 + settings 分 Tab

**目标**：
1. **删除** `pf-upload-zone / pf-result-card / pf-auth-card / pf-primary-button` magic class；`pf-site-logo` / `pf-public-glow` / `pf-console-shell` / `pf-console-content` 改用 CSS 变量或保留为布局 utility
2. **删除** `theme_config.public.upload_style / card_style / button_style` 三个字段
3. **改写** `SiteThemeRuntime`：唯一职责是解析合并后的 theme + user_override 为 CSS 变量写进 `:root`
4. **拆** `settings-page.tsx` 为 4 个 panel

**`SiteThemeRuntime` 新职责（伪代码）**：

```ts
function useComputedCssVariables() {
  const { site, userOverride } = usePersonalization()
  return useMemo(() => `
    :root {
      --pf-density: ${densityValue(site, userOverride)};        /* 0.75 | 1 | 1.25 */
      --pf-motion-duration: ${motionValue(site, userOverride)}; /* 0ms | 150ms | 300ms */
      --pf-public-glow-opacity: ${site.backgroundStyle === 'image' ? 0.45 : 1};
      --pf-logo-radius: ${logoRadius(site.logoShape)};
    }
  `, [site, userOverride])
}

export function SiteThemeRuntime() {
  const css = useComputedCssVariables()
  return <style id="picfast-site-theme">{css}</style>
}
```

**index.css 新写法**（density/motion 全局生效）：

```css
:root {
  --pf-density: 1;
  --pf-motion-duration: 150ms;
  --pf-public-glow-opacity: 1;
  --pf-logo-radius: 0.75rem;
}

/* 组件按需引用，无需 opt-in */
.pf-public-glow { opacity: var(--pf-public-glow-opacity); }
.pf-site-logo   { border-radius: var(--pf-logo-radius); }
```

组件侧使用方式（举例，**非强制**——需要密度的组件才写）：

```tsx
<div className="flex gap-3" style={{ gap: 'calc(0.75rem * var(--pf-density))' }}>
<button className="transition-colors" style={{ transitionDuration: 'var(--pf-motion-duration)' }}>
```

**文件改动**：

| 文件 | 改动 |
|---|---|
| `web/src/lib/theme-config.ts` | `ThemeConfig.public` 删除 `upload_style / card_style / button_style`；`themePresets` 6 套同步删；`mergeThemeConfig` 不再处理这 3 项 |
| `web/src/lib/theme-package.ts` | `assertSafeString` / `parsePublic` 删除这 3 项的解析；`THEME_PACKAGE_KEYS` 不变（仍含 `public`） |
| `web/src/components/site-theme-runtime.tsx` | 重写：去 dataset 写入，改写 CSS 变量 |
| `web/src/index.css` | 删 `data-pf-upload-style / data-pf-card-style / data-pf-button-style` 规则块；保留 `data-pf-motion / data-pf-density` 但只用来驱动 CSS 变量值；`pf-public-glow` / `pf-site-logo` 改用变量 |
| `web/src/components/upload-zone.tsx` | 删 `pf-upload-zone` class；密度相关 padding 改用 CSS 变量 |
| `web/src/components/upload-result.tsx` | 删 `pf-result-card` / `pf-primary-button` class |
| `web/src/pages/public/login-page.tsx` | 删 `pf-auth-card` class |
| `web/src/pages/layouts/public-layout.tsx` | `pf-public-glow` 改用 CSS 变量（可能直接删 class） |
| `web/src/pages/layouts/console-layout.tsx` | `pf-site-logo` 改用 CSS 变量；`pf-console-shell` / `pf-console-content` 评估是否仍需 |
| `web/src/pages/console/admin/settings/appearance-settings-page.tsx` | 删 `theme_background_style / theme_logo_shape` 之外的视觉轴字段；预览面板同步去 card/upload/button 三态 |
| `web/src/pages/console/settings-page.tsx` | 拆为 `settings/` 目录：`account-panel.tsx` / `workflow-panel.tsx` / `output-panel.tsx` / `appearance-panel.tsx`，外层用 Tabs；新增 appearance-panel 用 `<UserAppearanceEditor>` |
| `web/src/components/user-appearance-editor.tsx` (new) | 复用 admin 端 `<SiteThemeEditor>` 的 preset 卡片 + 简化字段（仅 preset/mode/density/motion + 跟随站点开关） |

**不引入**：`useThemeStyle` hook、`<Surface>` 组件、`<ThemeSurface>` 等中间抽象层——CSS 变量 + 原生 style 属性已足够。

**验证**：
- 6 个预设切换 → 所有相关组件都跟着变（不再"换了一半"）
- density 切到 compact → 表格行/列表项/上传区 padding 全局变密
- motion 切到 playful → 所有 transition 全局变长
- magic class 全部 grep 不到（除布局 utility）
- 6 个 preset 序列化/反序列化仍正确（删字段后 round-trip 通过）
- 后端 `validateThemeConfig` 同步删这 3 项的白名单检查

**提交**（拆 2 个）：
1. `refactor(theme): remove upload/card/button style axes, drive density/motion via CSS variables`
2. `refactor(settings): split settings page into typed panels`

### 6.4 v3：用户主题包导入导出（可选）

只动 `web/src/lib/theme-package.ts` + 新增 `<UserThemePackage />` 区块在 user appearance 面板。后端 `validateUserSettings` 加 `theme_override` 的导入路径校验。

**提交**：
1. `feat(theme): support user-scoped theme package import/export`

## 7. 兼容性分析与风险

### 7.1 数据库现状

- `site_settings`（单行 id=1）：`theme_config JSONB`、`default_copy_format`、`copy_template`、`allow_user_image_processing`、`skip_image_processing`。零约束在数据库层，校验只在 handler 入口。
- `users.settings`（每行 JSONB）：当前声明 6 个字段（`DefaultAlbum/DefaultStrategy/DefaultPermission/ImageProcessing/DefaultCopyFormat/CopyTemplate`），后端 `uploadUserSettings` 实际只读 `default_strategy` + `image_processing`——这是已存在的 schema/code 漂移。
- `groups.configs`：策略/配额，不进入本体系。
- localStorage：`theme`（next-themes）、`i18nextLng`、`default_strategy_id` / `default_album_id` / `default_permission`。

### 7.2 风险与对策

| 风险 | 影响 | 对策 |
|---|---|---|
| 后端 `validateThemeConfig` 白名单锁死新 key | 新增 theme_config 顶层 key 时 admin 保存 400 | v0 扩白名单；v2 改成"未知 key 警告但不拒绝"或参考 `analytics_config` 透传 |
| `UpdateProfile` 整体覆盖 | 新加 `theme_override` 时旧键被清空 | `usePersonalization()` 内统一 merge；提供 `patchUserSettings` helper |
| `UserSettings` Go struct 与 TS 类型漂移 | 后端读不到前端写入的字段 | v1 `validateUserSettings` 按 struct 全量校验，一次性消除漂移 |
| localStorage 旧值丢失 | 旧设备首次切到服务端后默认值看似"消失" | 双读（server ?? localStorage）+ 失败保留 + `pf-migrated-v1` 标志（见 §6.2） |
| localStorage 迁移网络失败 | 偏好看似丢 → 用户投诉 | 失败时保留 localStorage，下次挂载重试，toast 提示（见 §6.2） |
| `site_settings.theme_config.custom_css` XSS | 扩 CSS 注入范围会放大风险 | Go + TS 双侧维持 `;{}<>` 黑名单 + 20000 字符上限 |
| magic class 删不干净 | 视觉轴残留，回归 | v2 提交前 grep `pf-(upload-zone|result-card|auth-card|primary-button)` 必须 0 命中 |
| 多实例/多主题版本共存 | 老数据可能不识别新 token | `mergeThemeConfig` 已有 preset 兜底机制，**天然向前兼容**；重命名时需写一次性数据迁移 |

### 7.3 兼容保证

| 改造点 | 触碰面 | 是否破坏兼容 |
|---|---|---|
| v0 `personalization.ts` + hook 重构 | 仅前端 | **不破坏** |
| v0 扩 `ThemeTokenSet`（审计驱动） | 仅前端 + preset 数据扩展 | **不破坏**（后端白名单不限制 token 子键，缺失值由 `mergeThemeConfig` 兜底） |
| v1 新增 `users.settings.theme_override` / `language` 键 | 前端 + 后端 | **不破坏**（加 key 不删 key） |
| v1 `validateUserSettings` 全量校验 | 后端 | **不破坏**（校验只拒绝未知 key 和错误类型，不修改存量数据） |
| v1 `uploadUserSettings` 扩字段 | 后端 | **不破坏**（旧值不存在时回退到 group / admin 默认） |
| v1 localStorage → server 迁移 | 前端 | **不破坏**（双读 fallback + 失败保留 + 迁移标志） |
| v2 删除 `public.upload/card/button_style` + magic class | 前端 + 端到端 | **不破坏**（6 个 preset 的存量 JSON 会带这些字段，但前端 `mergeThemeConfig` 不再处理，后端白名单删除；老 theme 包导入时多余字段被忽略） |
| v2 改用 CSS 变量驱动 density/motion | 仅前端 | **不破坏**（旧组件读 `data-pf-density` 仍能拿到值，新增 CSS 变量兼容） |
| v2 settings Tab 化 | 仅前端 | **不破坏** |
| v3 user theme package | 前端 + 后端 `validateUserSettings` | **不破坏** |

**全程 0 条 migration**：`site_settings` 已经够用（`theme_config` 是 JSONB），`users.settings` 同样是 JSONB。新增字段都是"加 key 不删 key"。

### 7.4 回滚策略

每个 PR 独立可回滚：

- v0 revert：所有 hook 重构 revert 后，前端读 `localStorage` 行为完全不变
- v1 revert：去掉 `theme_override` 读写后，`mergeThemeConfig(site, undefined) = site` 行为不变；localStorage 兜底仍在
- v2 revert：恢复 `pf-*` class + `data-pf-*-style` 规则块 + `theme_config.public` 三个字段，前端 v2 改动 revert 后视觉回到 v1 状态
- v3 revert：去掉 user theme package 入口，保留 v2 行为

## 8. 与现有文档的关系

- `docs/theme-system.md`：保留，作为 theme 系统的"基线说明"。本文档是"基线 + 用户层 + 矩阵化"的扩展设计。
- `README.md`：本设计实施完成后，更新"特性"章节提及 5 个 type 的个人偏好 + 主题包导出。
- `CHANGELOG.md`：按 release-please 自动生成，但本设计的 v0/v1/v2 应拆成多个 feat/refactor commit 让版本号自然 bump。

## 9. 开放问题

1. **用户能不能改 token**（颜色）？当前设计为不能（`theme_override` 只含 preset/mode/density/motion）。若想开放，是否限制为"6 个核心 token 滑块"？待评审。
2. **language 是否要 server 持久化**？登录用户希望跨设备同步，未登录访客用 localStorage。v1 阶段建议保留双通道。
3. **用户主题包分享的范围**：是否允许分享到外部 URL？v3 阶段决定；当前不做。
4. **删除的 3 项（upload/card/button_style）是否真的不需要兜底导出**？admin 极端情况下想批量改"全站卡片变 glass"，目前只能写 custom_css。是否值得加一个"高级样式集"白名单？v3 之后再评估。
