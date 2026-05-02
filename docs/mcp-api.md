# PicFast MCP 接入指南

PicFast 通过本地 MCP 包提供完整的 AI 集成能力，支持图片上传和管理。

## 安装

```bash
npx @picfast/mcp
```

不需要全局安装，`npx` 会自动拉取并运行。

## 配置

在 MCP 客户端（Claude Desktop / Cursor / VS Code 等）中添加：

```json
{
  "mcpServers": {
    "picfast": {
      "command": "npx",
      "args": ["-y", "@picfast/mcp"],
      "env": {
        "PICFAST_BASE_URL": "https://your-picfast.example.com",
        "PICFAST_API_TOKEN": "img_xxxxxxxx"
      }
    }
  }
}
```

**环境变量：**

| 变量 | 必填 | 说明 |
|------|------|------|
| `PICFAST_BASE_URL` | 是 | PicFast 服务地址 |
| `PICFAST_API_TOKEN` | 否 | API Token，在 PicFast 控制台创建（以 `img_` 开头） |

## 可用工具

### `upload_image`

上传本地图片文件。

**参数：**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_path` | string | 是 | 本地图片文件路径 |
| `filename` | string | 否 | 覆盖文件名，缺省取路径 basename |
| `album_id` | number | 否 | 目标相册 ID |
| `permission` | number | 否 | 0=私有，1=公开 |

**支持格式：** jpg, jpeg, png, gif, webp, bmp, svg, ico, tiff, tif, avif

**返回示例：**

```json
{
  "key": "abc123",
  "url": "https://picfast.example.com/i/abc123.png",
  "markdown": "![demo.png](https://picfast.example.com/i/abc123.png)",
  "html": "<img src=\"https://picfast.example.com/i/abc123.png\" alt=\"demo.png\" />",
  "bbcode": "[img]https://picfast.example.com/i/abc123.png[/img]",
  "mimetype": "image/png",
  "original_size": 58213,
  "width": 1200,
  "height": 900
}
```

### `list_images`

分页查询图片。参数：`page`（页码）、`page_size`（每页数量）。

### `get_image`

按 key 获取图片详情。参数：`key`（图片 key）。

### `delete_image`

删除图片（不可逆）。参数：`key`（图片 key）。

### `get_usage_stats`

查询存储用量和配额。无参数。

## Resources

| URI | 说明 |
|-----|------|
| `picfast://user/profile` | 当前用户信息和容量 |
| `picfast://images/{key}` | 按 key 获取图片详情 |

## 错误码

| 错误码 | 说明 |
|--------|------|
| `UNAUTHORIZED` | 无用户上下文或 token 无效 |
| `FORBIDDEN_SCOPE` | 缺少必需 scope |
| `IMAGE_NOT_FOUND` | 图片不存在或非本人资源 |
| `FILE_NOT_FOUND` | 本地文件路径不存在 |
| `INVALID_FILE_TYPE` | 不支持的图片格式 |
| `UPLOAD_FAILED` | 服务端上传失败 |
| `INTERNAL_ERROR` | 内部错误 |

## REST API 直传

如需从脚本或其他程序上传：

```
POST /api/v1/images
Content-Type: multipart/form-data
Authorization: Bearer <API_TOKEN>

字段：file (必填), album_id (可选), permission (可选)
```