# @picfast/mcp

English | [中文](https://github.com/atbeta/picfast/blob/main/mcp/README.zh-CN.md)

PicFast MCP Server — upload images and manage your image hosting via AI with local file access.

## Features

- **Upload images** from local file paths (no base64 overhead)
- **List, get, delete** images in your account
- **Usage stats** — check storage quota and image count
- **Pipeline status** — inspect image processing stages (upload, compression, thumbnail, moderation)
- **Resources** — access user profile and image details via MCP resources
- **Guest upload** — works without API token for anonymous uploads (requires the PicFast instance to have guest upload enabled)

## Installation

No global install needed. Run directly with `npx`:

```bash
npx @picfast/mcp
```

## Configuration

Set environment variables in your MCP client config:

| Variable | Required | Description |
|----------|----------|-------------|
| `PICFAST_BASE_URL` | Yes | Your PicFast server URL (e.g. `https://picfast.example.com`) |
| `PICFAST_API_TOKEN` | No | API token for authenticated operations. Omit for guest upload only. Create one in the PicFast console (tokens start with `img_`). |

### Claude Desktop

```json
{
  "mcpServers": {
    "picfast": {
      "command": "npx",
      "args": ["-y", "@picfast/mcp"],
      "env": {
        "PICFAST_BASE_URL": "https://picfast.example.com",
        "PICFAST_API_TOKEN": "img_xxxxxxxx"
      }
    }
  }
}
```

### Cursor / VS Code

Same config format as Claude Desktop in the respective MCP settings panel.

## Tools

| Tool | Requires Token | Description |
|------|:-:|-------------|
| `upload_image` | No | Upload a local image file |
| `list_images` | Yes | List your images with pagination |
| `get_image` | Yes | Get image details and share links |
| `delete_image` | Yes | Delete an image by key |
| `get_usage_stats` | Yes | Get storage quota and usage |
| `get_image_pipeline` | Yes | Get image processing pipeline status |

## Resources

| URI | Requires Token |
|-----|:-:|
| `picfast://user/profile` | Yes |
| `picfast://images/{key}` | Yes |

## Guest Mode

Without `PICFAST_API_TOKEN`, only `upload_image` is available. Uploaded images are stored anonymously with no ownership tracking. Guest upload must be enabled on the PicFast instance (admin settings → allow guest upload). If guest upload is disabled, you will receive an error — in that case, set `PICFAST_API_TOKEN` to authenticate.

## License

MIT