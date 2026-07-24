# picfast

Command-line upload client for [PicFast](https://github.com/atbeta/picfast).

## Quick start

```bash
# one-time setup (saved under your OS user config dir)
npx picfast config set url https://your-instance.com
npx picfast config set token your-api-token   # optional

npx picfast upload image.png
npx picfast upload --markdown image.png
npx picfast upload *.png
```

Environment variables still work and override the config file:

```bash
export PICFAST_URL=https://your-instance.com
export PICFAST_TOKEN=your-api-token   # optional
```

## Usage

```
picfast upload [flags] <file...>
picfast config set url <url>
picfast config set token [token]
picfast config set token --stdin
picfast config show
picfast config unset <url|token>

Environment:
  PICFAST_URL         Base URL (overrides config file)
  PICFAST_TOKEN       API token (overrides config file)
  PICFAST_CONFIG_DIR  Override config directory

Flags:
  --markdown     output markdown image link instead of URL
  --format       output format: url, markdown, html, bbcode
```

Config file path:

- macOS: `~/Library/Application Support/picfast/config.json`
- Linux: `~/.config/picfast/config.json`
- Windows: `%APPDATA%\picfast\config.json`

## Install globally

```bash
npm install -g picfast
```

Or use `npx picfast ...` without installing.

## License

MIT
