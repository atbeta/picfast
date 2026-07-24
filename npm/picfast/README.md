# picfast

Command-line upload client for [PicFast](https://github.com/atbeta/picfast).

## Quick start

```bash
export PICFAST_URL=https://your-instance.com
export PICFAST_TOKEN=your-api-token   # optional

npx picfast upload image.png
npx picfast upload --markdown image.png
npx picfast upload *.png
```

## Usage

```
picfast upload [flags] <file...>

Environment:
  PICFAST_URL    Base URL of your PicFast instance (required)
  PICFAST_TOKEN  API token for authenticated uploads (optional)

Flags:
  --markdown     output markdown image link instead of URL
  --format       output format: url, markdown, html, bbcode
```

## Install globally

```bash
npm install -g picfast
```

Or use `npx picfast upload ...` without installing.

## License

MIT
