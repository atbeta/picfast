# PicFast MCP API 约定

本文档定义 PicFast MCP Server 的稳定返回结构，供客户端（Cursor、Claude Desktop、自定义 MCP 客户端）做可靠解析。

## 1. 鉴权

- 使用 HTTP Header：
  - `Authorization: Bearer <API_TOKEN>`
- token 建议使用 API Token 管理页创建的令牌。
- 对于“永不过期”令牌，服务端会返回可用的 MCP `expiration` 值，避免客户端校验失败。

## 2. 成功响应约定

目前 MCP 工具以 `text content` 承载 JSON 字符串。客户端应将文本解析为 JSON。

### 2.1 `upload_image`

成功响应（示例）：

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

字段说明：

- `original_size`：客户端上传的原始字节数（base64 解码后）
- `stored_size`：服务端最终写入存储的字节数
- `processed`：
  - `true`：发生了压缩、转码、水印等处理
  - `false`：原样存储（或处理失败后回退原始数据）

## 3. 错误响应约定

所有工具错误统一返回：

```json
{
  "error": {
    "code": "UPLOAD_FAILED",
    "message": "upload failed: file extension .exe is not allowed"
  }
}
```

## 4. 错误码表（当前最小集合）

- `UNAUTHORIZED`
  - 无有效用户上下文，或 token 不可用
- `FORBIDDEN_SCOPE`
  - token 缺少必需 scope（例如 `write`）
- `INVALID_IMAGE_DATA`
  - 上传参数格式错误（例如 `file_data` 不是合法 base64）
- `UPLOAD_FAILED`
  - 上传流程业务失败（格式、限额、策略、容量等）
- `IMAGE_NOT_FOUND`
  - 目标图片不存在，或无权访问该图片
- `INTERNAL_ERROR`
  - 服务内部异常

## 5. 客户端兼容建议

- 优先按 JSON 解析 `text content`，不要依赖纯文本句式。
- 错误处理应先读 `error.code`，再展示 `error.message`。
- 对 `upload_image`，建议展示 `processed` 与大小变化（`original_size -> stored_size`），便于排障与可观测。
