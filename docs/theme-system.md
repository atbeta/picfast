# PicFast Theme System

PicFast 0.8 introduces site-level appearance configuration for lightweight personalization.

## Current Scope

The theme system is intentionally token-based:

- Built-in presets define the visual baseline.
- Admin settings can override core colors, radius, public-page background, logo shape, and custom CSS.
- Public `/api/v1/config` returns `theme_config` so unauthenticated pages can render the selected theme.
- The frontend applies themes through CSS variables, keeping existing Tailwind/shadcn classes intact.

## Theme Config Shape

```json
{
  "preset": "moe",
  "mode": "system",
  "tokens": {
    "light": {
      "primary": "oklch(0.68 0.18 345)",
      "accent": "oklch(0.9 0.07 20)",
      "radius": "1rem"
    },
    "dark": {
      "primary": "oklch(0.74 0.18 345)",
      "accent": "oklch(0.36 0.09 20)",
      "radius": "1rem"
    }
  },
  "public": {
    "background_image": "",
    "background_style": "soft",
    "logo_shape": "circle"
  },
  "custom_css": ""
}
```

## Built-In Presets

- `default`: neutral PicFast style.
- `moe`: soft anime-inspired colors and rounder surfaces.
- `cyber`: dark-first neon styling.
- `pixel`: low-radius retro style.
- `terminal`: green terminal-inspired style.
- `fresh`: light cyan/mint style.

## Future Local Theme Packages

A future version can support local file-based themes without introducing a marketplace or remote installer.

Recommended shape:

```text
data/themes/sakura/
  theme.json
  theme.css
  preview.png
```

`theme.json` should map to the same `theme_config` schema:

```json
{
  "id": "sakura",
  "name": "Sakura",
  "version": "1.0.0",
  "author": "local",
  "config": {
    "preset": "sakura",
    "mode": "system",
    "tokens": {
      "light": {
        "primary": "oklch(0.68 0.18 345)"
      }
    },
    "public": {
      "background_image": "./bg.png",
      "background_style": "image"
    }
  },
  "css": "./theme.css",
  "preview": "./preview.png"
}
```

Keep local themes CSS-only. Avoid executable JavaScript or React plugins so self-hosted instances remain easy to upgrade and reason about.
