# PicFast Theme System

> 状态：当前实现（2026-06）
> 升级指南：见 `docs/theme-migration.md`（针对之前使用预设 / token / 主题包功能的用户）

## 1. 设计目标

站点使用**单一内置主题**，admin 通过 `theme_config.custom_css` 注入细节定制。这是一个 2026 年的设计简化，砍掉了早期版本里的 6 套预设、token 调色板、JSON 主题包导入导出、用户级 `theme_override`。

简化理由（完整的设计演进见 `docs/personalization.md` §3.2）：

- 6 套预设里大多数部署只用 1 套，"换预设"的实际触发率极低
- token 调色板（primary / accent / radius 等）由前端组件直接读 CSS 变量即可，无需 admin 改 JSON
- 用户级 `theme_override` 的 4 个轴（preset / mode / density / motion）几乎没人配
- 想要差异化细节时，custom_css 已经能覆盖

## 2. 数据模型

### 2.1 `site_settings.theme_config` JSONB 字段

```jsonc
{
  "mode": "system",      // light | dark | system（可选；未登录访客的兜底）
  "custom_css": ""        // 注入到站点 <style> 标签的 CSS（可选）
}
```

- 顶层只允许这两个 key（后端 `validateThemeConfig` 严格白名单）
- `custom_css` 长度上限 20000 字符
- `mode` 只在用户未登录或用户本地未设置时生效（next-themes 优先）

### 2.2 不再支持的字段

之前的 `preset` / `tokens.{light,dark}.*` / `public.{upload_style, card_style, button_style, background_*, logo_shape}` 全部不再读取，存量数据保留为死值。JSONB 加 key 不删 key，所以升级不需要数据迁移。

## 3. 注入点

admin 站点设置页（`/console/admin/site`）的"自定义 CSS"section：

```
┌─────────────────────────────────────────────┐
│  自定义 CSS                                  │
│  保存后会原样注入到站点 <style> 标签中...   │
│  ┌─────────────────────────────────────┐    │
│  │                                     │    │
│  │  :root { --primary: ... }          │    │
│  │                                     │    │
│  └─────────────────────────────────────┘    │
│  可优先覆盖 .pf-public-glow / .pf-site-logo │
│  / .pf-console-shell / .pf-console-content │
│  也可以使用 var(--primary) 等主题变量。     │
└─────────────────────────────────────────────┘
```

实现：`web/src/pages/console/admin/settings/site-settings-page.tsx`（合并后）+ `web/src/components/site-theme-runtime.tsx`（渲染层）。

## 4. 渲染时序

`SiteThemeRuntime` 是个全局挂载的 React 组件，挂在 React tree 根上：

```ts
const css = (site?.theme_config as { custom_css?: string } | null)?.custom_css ?? ''
// ...
return <style id="picfast-site-theme">{customCSS}</style>
```

CSS 变量和 keyframes 通过浏览器原生 `<style>` 注入，加载顺序：

1. `web/src/index.css`（内置主题的 `:root` 变量 + Tailwind base）
2. 用户的 `custom_css`（最后追加，CSS 级联优先级最高）

## 5. 可覆盖的钩子

### 5.1 公开类（CSS 选择器）

| 选择器 | 用途 | 备注 |
|---|---|---|
| `.pf-public-glow` | 公开页背景光晕 | 透明度由 CSS 变量 `--pf-public-glow-opacity` 控制（保留为 1，需要关掉时直接 `display: none`） |
| `.pf-site-logo` | 站点 logo 形状 | 圆角由 `--pf-logo-radius` 控制（默认 `0.75rem`） |
| `.pf-console-shell` | 控制台整体 grid 容器 | — |
| `.pf-console-content` | 控制台主内容卡片 | — |

### 5.2 主题变量（CSS variables）

最常用的几个，全部在 `:root` 定义：

```css
--background, --foreground
--card, --card-foreground
--primary, --primary-foreground
--secondary, --secondary-foreground
--muted, --muted-foreground
--accent, --accent-foreground
--destructive, --destructive-foreground
--warning, --warning-foreground
--success, --success-foreground
--info, --info-foreground
--border, --input, --ring
--popover, --popover-foreground
--chart-1, --chart-2, --chart-3, --chart-4, --chart-5
--sidebar, --sidebar-foreground, --sidebar-primary, --sidebar-primary-foreground
--sidebar-accent, --sidebar-accent-foreground, --sidebar-border, --sidebar-ring
--pf-density: 1;                    /* 静态，无需改 */
--pf-motion-duration: 150ms;        /* 静态，无需改 */
--pf-logo-radius: 0.75rem;          /* 可被 custom_css 覆盖 */
```

完整列表见 `web/src/index.css`（`@theme inline` 块附近）。`.dark` 选择器下的同名变量会自动在深色模式下生效。

### 5.3 静态层（无需改）

`--pf-density` 和 `--pf-motion-duration` 是 v0.16 收敛时硬编码在 `:root` 的全局常量，不再通过 theme_config 暴露。如果实例想做密度 / 动效调整，**用 custom_css 覆盖即可**：

```css
:root {
  --pf-density: 0.85;            /* 比默认紧 15% */
  --pf-motion-duration: 100ms;
}
```

## 6. 常见用法

### 6.1 改主色

```css
:root {
  --primary: oklch(0.62 0.18 280);
  --ring: oklch(0.62 0.18 280);
}
```

### 6.2 改 logo 形状

```css
.pf-site-logo {
  border-radius: 9999px;     /* 圆形 */
}
```

### 6.3 全局加阴影

```css
.pf-console-content {
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}
```

### 6.4 隐藏背景光晕

```css
.pf-public-glow { display: none; }
```

## 7. 模式切换

`theme_config.mode` 的语义是**未登录访客的兜底**，登录用户由 `next-themes` 本地选择接管。

```
未登录访客：
  site.theme_config.mode === 'dark'  → 站点 dark
  site.theme_config.mode === 'light' → 站点 light
  site.theme_config.mode === 'system' → 跟随系统 prefers-color-scheme

登录用户：
  user 本地 theme（localStorage）优先
  未设置 → 回退到 site.theme_config.mode
```

UI 入口在 header 右侧的 `ThemeSwitcher`（`web/src/components/theme-switcher.tsx`），提供 light / dark / system 三态切换，写 localStorage，不写 server。

## 8. 不再支持的功能

| 旧功能 | 替代方案 | 备注 |
|---|---|---|
| 6 套主题预设 | — | 已删除；选 default 风格 |
| 调色板 token（primary / accent / radius 等覆盖） | custom_css 写 `:root` | 推荐用法见 §6 |
| 主题包导入导出（`{ preset, mode, tokens, public, custom_css }`） | 手工编辑 custom_css | 单站点配置没必要做包 |
| 用户级 `theme_override`（`{ preset, mode, density, motion }`） | header 切换器 + custom_css | 4 个轴 → 1 个 |
| 密度 / 动效可配置档（compact/comfortable/spacious × none/subtle/playful） | 改 `--pf-density` 和 `--pf-motion-duration` 变量 | 3 档 → 任意连续值 |

迁移方法：见 `docs/theme-migration.md`。
