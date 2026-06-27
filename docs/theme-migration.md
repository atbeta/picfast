# 主题升级指南（v0.16）

> 适用对象：升级到 PicFast 0.16+ 的现有部署
> 范围：仅涉及"主题 / 视觉" 相关的 breaking change
> 兼容性：所有数据保留为 JSONB 死值，**不需要任何数据迁移**

## 0. 发生了什么

v0.16 把 6 套主题预设、token 调色板、JSON 主题包导入导出、用户级 `theme_override` **整体砍掉**。原因是实际部署里这几个功能几乎没人用，维护成本却很高。

升级后：

- 站点使用**单一内置主题**
- admin 通过 `theme_config.custom_css` 注入细节定制
- 普通用户除了 header 里的亮/暗/系统切换，没有其他主题相关设置
- 之前的所有 `theme_config` / `users.settings.theme_override` 数据**原样保留**，只是不再被读取

详见 `docs/theme-system.md`（新模型）和 `docs/personalization.md`（设计演进）。

## 1. 影响清单

### 1.1 admin 站点设置

- ❌ 移除：「外观风格」独立页面（被合并到「站点信息」的"自定义 CSS" section）
- ❌ 移除：「主题包」导入导出
- ✅ 保留：「站点信息」+「自定义 CSS」+「统计集成」（合并后 1 个页面）

**升级前** 你可能选了 `moe / cyber / pixel / terminal / fresh` 之一，或在 token 区域写了 `primary: oklch(...)`、`radius: 1rem` 等。

**升级后** 这些值都还在 `site_settings.theme_config` JSONB 里（数据库层不删），但**应用不再读取**。视觉上站点会回到 default 主题。

### 1.2 用户设置

- ❌ 移除：「个性化 → 外观」整个 tab（用户级主题覆盖）
- ❌ 移除：用户级 `theme_override.preset / mode / density / motion`
- ✅ 保留：header 里的亮/暗/系统切换（这是基础设施，不是"个性化"）

**升级前** 你的 user 可能设过 `theme_override.preset = "moe"` 或 `density = "compact"`。

**升级后** migration `029_cleanup_theme_override` 会**一次性清掉**这些 `theme_override` 字段。用户的 mode 选择回退到 admin 站点的 `theme_config.mode`，density/motion 回到 CSS 变量硬编码的默认值。

### 1.3 数据库 schema

| 改动 | 是否 breaking | 备注 |
|---|---|---|
| `site_settings.theme_config` 字段保留 | ❌ 不破坏 | 之前的 `preset` / `tokens.*` / `public.*` 值仍存在，只是不读 |
| `users.settings.theme_override` 字段 | ❌ 不破坏 | migration 029 主动清掉，但即使不跑 migration 也不影响功能（应用不读） |
| `theme_config` 字段类型 | ❌ 不破坏 | 仍是 JSONB，新白名单只允许 `mode` + `custom_css` |
| 新增 `theme_config.custom_css` 支持 | ✅ 增强 | 之前没有这个字段，是新加的 |

## 2. 迁移步骤

### 2.1 应用迁移（必须）

按标准 PicFast 升级流程部署 0.16+ 镜像。

```bash
# 标准升级
docker compose pull
docker compose up -d

# 数据库迁移会自动跑（migration 029 会清掉用户的 theme_override）
```

### 2.2 把旧主题设置转成 custom_css（推荐）

**如果你之前选了某个预设**（如 `moe`），需要在 admin 站点设置的"自定义 CSS"里手动补回对应的视觉效果。下面的表给出每个预设的近似恢复方式：

#### `default` 预设
无需任何 custom_css。

#### `moe` 预设（粉系柔光）
```css
:root {
  --primary: oklch(0.70 0.16 350);          /* 粉色高亮 */
  --accent: oklch(0.92 0.05 350);
  --background: oklch(0.99 0.01 350);
  --ring: oklch(0.70 0.16 350);
}
.pf-site-logo { border-radius: 1.25rem; }    /* 更圆润 */
```

#### `cyber` 预设（赛博霓虹）
```css
:root {
  --background: oklch(0.13 0.02 280);
  --foreground: oklch(0.95 0.02 280);
  --card: oklch(0.18 0.03 280);
  --primary: oklch(0.78 0.20 200);          /* 青色 */
  --accent: oklch(0.72 0.25 320);           /* 品红 */
  --border: oklch(0.28 0.04 280);
  --ring: oklch(0.78 0.20 200);
}
```

#### `pixel` 预设（像素乐园）
```css
:root {
  --radius: 0.125rem;                       /* 硬圆角 */
  --border: oklch(0.85 0 0);
}
```

#### `terminal` 预设（黑绿终端）
```css
:root {
  --background: oklch(0.05 0 0);
  --foreground: oklch(0.85 0.18 145);
  --card: oklch(0.08 0 0);
  --primary: oklch(0.75 0.22 145);
  --border: oklch(0.20 0.05 145);
  --ring: oklch(0.75 0.22 145);
}
```

#### `fresh` 预设（清新海盐）
```css
:root {
  --primary: oklch(0.72 0.10 195);          /* 青色 */
  --accent: oklch(0.94 0.04 195);
  --background: oklch(0.99 0.005 195);
  --ring: oklch(0.72 0.10 195);
}
.pf-site-logo { border-radius: 0.875rem; }
```

**如果你之前在 token 区域写过 primary / accent / radius**，把对应的 `:root { --key: value; }` 复制到 custom_css 即可。

**如果你之前在 `public.background_image` 填过 URL**：
```css
body::before {
  content: '';
  position: fixed;
  inset: 0;
  z-index: -1;
  background-image: url('你的URL');
  background-size: cover;
  background-position: center;
}
.pf-public-glow { display: none; }            /* 关闭光晕 */
```

**如果你之前在 `public.logo_shape` 选过 circle / square**：
```css
.pf-site-logo { border-radius: 9999px; }      /* circle */
.pf-site-logo { border-radius: 0; }            /* square */
```

### 2.3 检查用户态偏好（推荐）

升级后，**每个用户** 都需要看一次他们的设置页。用户的 `theme_override` 已经被自动清掉，但 mode 行为有变化：

- **之前**：用户可以选 `light / dark / system`
- **现在**：还是 `light / dark / system`，但存的是 localStorage 而不是 server

如果你的部署中所有用户都依赖 server 端 mode 偏好，告知他们去 header 里切一下（localStorage 是浏览器级，跨设备不共享）。

## 3. 回滚方案

万一升级后站点样式不理想：

1. 立刻回滚到 v0.15 镜像（数据层不需要回滚，theme_config 还在）
2. 检查 `site_settings.theme_config` 的旧值
3. 提取旧 `tokens.{light,dark}` 和 `public.*` 字段
4. 按 §2.2 的转换表转成 custom_css
5. 重新升到 v0.16+，把 custom_css 贴进 admin 站点设置

## 4. FAQ

### Q: 我没设过任何 theme，会不会受影响？
不会。你的 `site_settings.theme_config` 是 `{}` 或 `{"custom_css": "..."}`，完全没影响。

### Q: 我的用户在他们的账号下选了 `cyber` 主题，会怎样？
用户视觉上回退到 default 主题。`theme_override` 字段被 migration 029 清掉。告知用户在 header 切到 dark mode 可以部分接近 cyber 风格（如果你的 default 主题偏暗）。

### Q: 升级前我可以导出我的配置吗？
可以。在升级前打开 `/console/admin/site`，把"外观风格"页面（v0.15 之前）的"主题包 → 复制 JSON"按钮拷一份下来作为备份。

但**注意**：v0.16 已经没有"主题包 → 导入"按钮了，备份的 JSON 你需要手工按 §2.2 转成 custom_css。

### Q: 砍掉这些功能后悔了想恢复怎么办？
恢复是不可能的（迁移是单向的）。但你可以：
- 用 custom_css 重建 6 套预设的效果（见 §2.2 的转换表）
- 提交 feature request，未来版本可能加回"预设库"（参考 docs-site 中的"Backup & Maintenance"模式，主题包可能以另一种形式回来）

### Q: 我看了 docs/theme-system.md 里的 CSS 变量列表，但有些变量我没找到
`:root` 下的所有变量都在 `web/src/index.css` 的 `@theme inline` 块附近，`.dark` 块有深色版本的同名变量。如果你想要某个不在列表里的 CSS 属性，可以通过 custom_css 直接写选择器样式。

### Q: 我的 custom_css 长度刚好到 20000 字符上限了怎么办？
v0.16 之后硬限制 20000 字符（v0.15 之前没有显式限制）。考虑：
- 合并重复规则
- 拆分成 `:root` 变量 + 实际使用处的引用
- 减少 vendor-prefix（PicFast 用 Tailwind 已经处理了兼容性）

## 5. 相关改动一览

| 改动 | 关联 commit | 影响范围 |
|---|---|---|
| 砍 6 套预设 | `278b77a` | admin 站点设置 |
| 砍 theme_override + users.settings cleanup | `278b77a` + `e0e8e03` | 用户设置 |
| 合并 appearance + analytics 到 site settings | `4850f5b` | admin 站点设置（nav） |
| 拆分 /console/settings → /console/account + 上传弹窗 | `3b95cf1` | 用户设置（路由） |
| 站点信息页 custom_css 改用 stack 布局 | `7d3eb72` | admin 站点设置（UI 细节） |
| 删 maintenance 页面 → docs site | `babd0d3` | admin nav |
| 删 doc 站 `themeConfig` 旧字段描述 | `ba7939c` | docs site |
| 补 image-processing 上传时处理段 | `d3bf23b` | docs site |
