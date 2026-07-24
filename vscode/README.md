# PicFast Image Uploader

Upload local images from VS Code to a self-hosted [PicFast](https://github.com/atbeta/picfast)
instance. Works with both guest uploads and API tokens.

- **`Ctrl+Alt+U`** / **`Cmd+Alt+U`** — scan the current markdown/HTML file, upload every
  local image reference, and replace paths in place
- **`Ctrl+Alt+V`** / **`Cmd+Alt+V`** — pick a single local image, upload it, insert the link
  at the cursor
- Recognizes `![alt](path)` and `<img src="path" />` references, skips remote URLs
- All replacements applied as a single undoable edit
- Status bar feedback and cancellable progress notification

## Requirements

- VS Code **1.78** or later
- A reachable PicFast instance. Authenticated uploads work with an API token;
  guest uploads require `allow_guest_upload: true` on the server.

## Quick start

1. Install the extension from Marketplace.
2. Run **"PicFast: Set Base URL"** from the Command Palette, or set
   `picfast.baseUrl` in settings. Example: `https://demo.picfast.dev`
3. Optionally set `picfast.apiToken` for authenticated uploads.
4. Open a markdown file, press **`Ctrl+Alt+U`**.
5. For one-off uploads, press **`Ctrl+Alt+V`** and pick a file.

## Commands

| Command | Title | Keybinding |
|---|---|---|
| `picfast.uploadLocalImages` | Upload and Replace Local Image References | `Ctrl+Alt+U` / `Cmd+Alt+U` |
| `picfast.pasteImage` | Upload Local Image File | `Ctrl+Alt+V` / `Cmd+Alt+V` |
| `picfast.setBaseUrl` | Set Base URL | — |
| `picfast.setApiToken` | Set API Token | — |
| `picfast.openDocs` | Open Documentation | — |

Keybindings can be changed via `Ctrl+K Ctrl+S` → search "picfast".

## Configuration

| Setting | Type | Default | Description |
|---|---|---|---|
| `picfast.baseUrl` | string | `https://demo.picfast.dev` | Base URL of your PicFast instance (no trailing slash) |
| `picfast.apiToken` | string | `""` | Optional API token; leave empty for guest uploads |
| `picfast.defaultFormat` | enum | `markdown` | Insert format: `url` / `markdown` / `html` / `bbcode` |
| `picfast.timeoutMs` | number | `30000` | HTTP upload timeout (1000–300000 ms) |
| `picfast.showStatusBar` | boolean | `true` | Show the status bar item |

## License

GPL-3.0-or-later
