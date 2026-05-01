# PicFast MCP API 参考

本文档对应当前服务端实现（`internal/handler/mcp_server.go`），用于 MCP 客户端做稳定接入和错误处理。

## 1) 鉴权与作用域

- 传输：`Authorization: Bearer <API_TOKEN>`
- Token 来源：控制台的 API Token 管理
- 默认 scopes：若创建 token 未传 `scopes`，默认是 `["read", "write"]`
- 作用域规则：
  - `read`：允许读取图片/统计/资源
  - `write`：允许上传和删除

## 1.1) Server Instructions（客户端握手文案）

当前 MCP 服务端会向客户端暴露以下 instructions（与代码保持一致）：

```text
PicFast MCP Server — manage image hosting via AI.

Available tools:
- upload_image: Upload an image and get shareable links (URL, Markdown, HTML, BBCode).
- list_images: List your uploaded images with pagination.
- get_image: Get details and all format links for a specific image by key.
- delete_image: Delete an image by key.
- get_usage_stats: Get your storage usage and quota.

You can also read resources:
- picfast://user/profile — current user info and capacity.
- picfast://images — list of images (as resource).
```

## 2) 通用返回规范

### 2.1 工具成功返回

- MCP `CallToolResult.IsError = false`
- `content[0].text` 为 JSON 字符串（`delete_image` 例外，返回普通文本）

### 2.2 工具错误返回

- MCP `CallToolResult.IsError = true`
- `content[0].text` 结构固定：

```json
{
  "error": {
    "code": "UPLOAD_FAILED",
    "message": "upload failed: file size exceeds maximum (2097152 bytes)"
  }
}
```

### 2.3 错误码

- `UNAUTHORIZED`：无用户上下文或 token 无效
- `FORBIDDEN_SCOPE`：缺少必需 scope（如 `write`）
- `INVALID_IMAGE_DATA`：`file_data` 非法（base64 解析失败）
- `UPLOAD_FAILED`：上传流程失败（容量/策略/格式/审核等）
- `IMAGE_NOT_FOUND`：图片不存在或非本人资源
- `INTERNAL_ERROR`：服务内部错误

## 3) Tool Reference

## 3.0) REST 对照关系（OpenAPI ↔ MCP）

- `upload_image` ↔ `POST /api/v1/images`
- `list_images` ↔ `GET /api/v1/images`
- `get_image` ↔ `GET /api/v1/images/{key}`
- `delete_image` ↔ `DELETE /api/v1/images/{key}`
- `get_usage_stats` ↔（组合查询，暂无单一 REST 等价端点）
- MCP 鉴权 token 创建：`POST /api/v1/api-tokens`

## `upload_image`

- **用途**：上传图片并返回可直接分享的链接
- **scope**：`write`
- **参数**：
  - `file_data` (string, required)：图片二进制的 base64 字符串
  - `filename` (string, required)：文件名（用于扩展名和展示名）
  - `album_id` (int64, optional)：目标相册 ID
  - `permission` (int16, optional)：权限级别（`0=private`, `1=public`）
- **成功返回示例**：

```json
{
  "key": "abc123",
  "url": "https://picfast.example.com/i/abc123.png",
  "markdown": "![demo.png](https://picfast.example.com/i/abc123.png)",
  "html": "<img src=\"https://picfast.example.com/i/abc123.png\" alt=\"demo.png\" />",
  "bbcode": "[img]https://picfast.example.com/i/abc123.png[/img]",
  "mimetype": "image/png",
  "original_size": 58213,
  "stored_size": 43102,
  "processed": true
}
```

## `list_images`

- **用途**：分页查询当前用户图片
- **scope**：`read`
- **参数**：
  - `page` (int32, optional, default 1)
  - `page_size` (int32, optional, default 20, max 100)
- **成功返回示例**：

```json
{
  "items": [
    {
      "key": "abc123",
      "origin_name": "demo.png",
      "size_bytes": 58213,
      "width": 1200,
      "height": 900,
      "url": "https://picfast.example.com/i/abc123.png",
      "permission": 1,
      "created_at": "2026-05-01T12:34:56Z"
    }
  ],
  "total": 42,
  "page": 1,
  "size": 20
}
```

## `get_image`

- **用途**：按 `key` 获取单图详情与多格式链接
- **scope**：`read`
- **参数**：
  - `key` (string, required)：图片唯一 key
- **成功返回示例**：

```json
{
  "key": "abc123",
  "origin_name": "demo.png",
  "size_bytes": 58213,
  "mimetype": "image/png",
  "width": 1200,
  "height": 900,
  "permission": 1,
  "url": "https://picfast.example.com/i/abc123.png",
  "thumbnail": "https://picfast.example.com/t/xxxx.png",
  "markdown": "![demo.png](https://picfast.example.com/i/abc123.png)",
  "html": "<img src=\"https://picfast.example.com/i/abc123.png\" alt=\"demo.png\" />",
  "bbcode": "[img]https://picfast.example.com/i/abc123.png[/img]",
  "created_at": "2026-05-01T12:34:56Z"
}
```

## `delete_image`

- **用途**：删除指定图片
- **scope**：`write`
- **参数**：
  - `key` (string, required)：图片唯一 key
- **成功返回**：纯文本 `"Image deleted successfully."`

## `get_usage_stats`

- **用途**：查询当前用户容量和数量统计
- **scope**：`read`
- **参数**：无
- **成功返回示例**：

```json
{
  "user_id": 1,
  "name": "admin",
  "capacity_bytes": 0,
  "used_bytes": 1048576,
  "image_num": 12,
  "album_num": 3
}
```

## 4) Resources Reference

## `picfast://user/profile`

- **scope**：`read`
- **返回示例**：

```json
{
  "id": 1,
  "name": "admin",
  "email": "admin@example.com",
  "role": "admin",
  "capacity_bytes": 0,
  "used_bytes": 1048576,
  "image_num": 12,
  "album_num": 3
}
```

## `picfast://images/{key}`

- **scope**：`read`
- **说明**：`{key}` 为图片 key，仅可读取本人图片
- **返回示例**：

```json
{
  "key": "abc123",
  "origin_name": "demo.png",
  "size_bytes": 58213,
  "width": 1200,
  "height": 900,
  "url": "https://picfast.example.com/i/abc123.png",
  "created_at": "2026-05-01T12:34:56Z"
}
```

## 5) Prompt

当前内置 prompt：

- `upload_and_share`
  - 语义：要求模型上传图片并返回 URL / Markdown / HTML / BBCode

## 6) 客户端实现建议

- 始终先尝试将 `content[].text` 解析为 JSON，再回退纯文本展示。
- 错误处理以 `error.code` 为主，`error.message` 直接展示给用户或记录日志。
- 对上传类场景展示 `original_size -> stored_size` 与 `processed`，便于理解压缩/处理行为。
- 如需调用管理能力（用户、分组、策略、审核），应走 OpenAPI 对应的 `/api/v1/admin/*` REST 接口。
