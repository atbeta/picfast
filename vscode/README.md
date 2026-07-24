# PicFast Image Uploader

A minimal VS Code extension that talks to a self-hosted
[PicFast](https://github.com/atbeta/picfast) instance. Ships with two
complementary commands:

1. **Upload and replace local image references** in the current
   markdown / HTML file (the workflow you usually want).
2. **Upload a single local file** (the escape hatch for non-markdown
   content or one-off uploads).

## Features

- **`Ctrl+Alt+U`** (or `Cmd+Alt+U` on macOS) — scan the current
  markdown/HTML file, upload every local image, and replace the local
  path with the remote URL in one go
- **`Ctrl+Alt+V`** (or `Cmd+Alt+V` on macOS) — pick a single local
  image, upload it, and insert the resulting link at the cursor
- Recognizes markdown `![alt](path)` (with optional title) and HTML
  `<img src="path" />` (single or double quoted)
- Skips non-local references (`http://`, `https://`, `data:`)
- Resolves relative paths against the document's directory; absolute
  paths and `file://` URIs work too
- Status bar feedback and a cancellable progress notification
- All replacements applied in a single `WorkspaceEdit` so the user sees
  one undo step
- Posts to `POST <baseUrl>/api/v1/flat/upload` (works with guest
  upload out of the box; no API token required)
- Inserts the resulting link in **URL / Markdown / HTML / BBCode** form

## Requirements

- VS Code **1.78** or later
- A reachable PicFast instance with `allow_guest_upload=true` (the
  default in `docker/.env.example`)

## Quick start

1. Install the extension (`.vsix` or from Marketplace).
2. Open the Command Palette and run **"PicFast: Set Base URL"**, or
   set `picfast.baseUrl` in VS Code settings.
   Example: `https://picfast.your-company.internal`
3. Open a markdown file that contains references like
   `![screenshot](./shots/login.png)`.
4. Press **`Ctrl+Alt+U`** (or `Cmd+Alt+U` on macOS). The extension
   uploads every local image it can resolve and replaces the markdown
   in place. A progress notification shows `1/N file.png`, the
   status bar reflects `$(check) PicFast: replaced N reference(s)`.

For one-off uploads outside a markdown file (e.g. into a config
snippet), use **`Ctrl+Alt+V`** and pick the file in the dialog.

## Commands

| Command | Title | Default keybinding |
|---|---|---|
| `picfast.uploadLocalImages` | PicFast: Upload and Replace Local Image References | `Ctrl+Alt+U` / `Cmd+Alt+U` |
| `picfast.pasteImage` | PicFast: Upload Local Image File | `Ctrl+Alt+V` / `Cmd+Alt+V` |
| `picfast.setBaseUrl` | PicFast: Set Base URL | — |
| `picfast.setApiToken` | PicFast: Set API Token | — |
| `picfast.openDocs` | PicFast: Open Documentation | — |

The `uploadLocalImages` keybinding is scoped to
`editorTextFocus && editorLangId =~ /^(markdown|mdx)$/`, so the
shortcut is a no-op in JSON / Go / etc. files.

## Customizing the keybindings

All keybindings are suggestions, not hard-coded. VS Code lets
you rebind any of them to whatever you prefer, per command,
without losing the `when` clause:

1. Press `Ctrl+K Ctrl+S` (macOS: `Cmd+K Cmd+S`) to open the
   Keyboard Shortcuts editor.
2. Type "picfast" in the search box. You'll see all the
   extension's commands with their current bindings.
3. Double-click the binding (or right-click → "Change
   Keybinding") and press the keys you want. The rebinding
   is scoped to whatever `when` clause the command carries
   (e.g. `editorTextFocus && editorLangId =~ /^(markdown|mdx)$/`
   for the bulk-upload shortcut), so it won't leak into
   other file types.
4. Optionally click the gear icon → "Add Keybinding" if you
   want a *secondary* binding alongside the default.

User-level rebinds live in `keybindings.json` (`File` →
`Preferences` → `Keyboard Shortcuts` → top-right `{}` icon) and
are preserved across extension updates.

## Configuration

| Setting | Type | Default | Description |
|---|---|---|---|
| `picfast.baseUrl` | string | `https://picfast.example.com` | Base URL of your PicFast instance (no trailing slash) |
| `picfast.apiToken` | string | `""` | Optional API token. Leave empty for guest uploads |
| `picfast.defaultFormat` | enum | `markdown` | `url` \| `markdown` \| `html` \| `bbcode` (pasteImage only) |
| `picfast.timeoutMs` | number | `30000` | HTTP upload timeout (1–300000 ms) |
| `picfast.showStatusBar` | boolean | `true` | Show the status bar item |

## How it talks to PicFast

```
POST {baseUrl}/api/v1/flat/upload
Content-Type: multipart/form-data
Body:        file=<png-bytes>

200 OK
{
  "url":          "https://.../i/abc.png",
  "thumbnail_url":"https://.../t/abc.png",
  "markdown":     "![name](https://.../i/abc.png)",
  "html":         "<img src=\"https://.../i/abc.png\" alt=\"name\" />",
  "bbcode":       "[img]https://.../i/abc.png[/img]"
}
```

This is the same `flat` upload endpoint that the `picfast` CLI uses,
so behaviour is identical across the two tools.

## How local references are detected

Two regular expressions drive the scan:

- `![alt](path)` and `![alt](path "title")` — markdown
- `<img ... src="path" ... />` (single or double quoted) — HTML

The scan then filters out any path that starts with `http://`,
`https://`, or `data:`, and resolves the remainder to an absolute
filesystem path against the document's directory. A reference is
uploadable if:

- the resolved file exists and is a regular file
- the extension is one of `png / jpg / jpeg / gif / webp / bmp / tif / tiff / avif / ico`

If the file is missing, not an image, or unreadable, that specific
reference is skipped and reported at the end of the run; other
references are unaffected.

## Build from source

```bash
cd vscode
npm install --include=dev    # the --include=dev flag is needed when
                             # NODE_ENV=production is set in your shell
npm run build
npm test
npm run package              # produces picfast-0.1.0.vsix
```

Install locally:

```bash
code --install-extension picfast-0.1.0.vsix
```

## License

GPL-3.0-or-later — same license as PicFast itself.
