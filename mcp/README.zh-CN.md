# @picfast/mcp

[English](https://github.com/atbeta/picfast/blob/main/mcp/README.md) | 中文

PicFast MCP 服务器 — 通过 AI 助手上传图片并管理你的图床，直接读取本地文件。

## 功能

- **上传图片** — 直接从本地路径读取，无需 base64 编码
- **列表、查看、删除** — 管理你账户中的图片
- **用量统计** — 查看存储配额与图片数量
- **资源协议** — 通过 MCP 资源访问用户信息和图片详情
- **游客上传** — 无需 API Token 即可匿名上传（需 PicFast 实例开启游客上传）

## 安装

无需全局安装，直接使用 `npx` 运行：

```bash
npx @picfast/mcp
```

## 配置

在 MCP 客户端配置中设置环境变量：

| 变量 | 必填 | 说明 |
|------|:----:|------|
| `PICFAST_BASE_URL` | 是 | PicFast 服务地址（如 `https://picfast.example.com`） |
| `PICFAST_API_TOKEN` | 否 | API Token，用于认证操作。不填则仅限游客上传。在 PicFast 控制台中创建（Token 以 `img_` 开头） |

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

在 MCP 设置面板中使用与 Claude Desktop 相同的配置格式。

## 工具

| 工具 | 需要 Token | 说明 |
|------|:----------:|------|
| `upload_image` | 否 | 上传本地图片文件 |
| `list_images` | 是 | 分页列出图片 |
| `get_image` | 是 | 获取图片详情与分享链接 |
| `delete_image` | 是 | 按键值删除图片 |
| `get_usage_stats` | 是 | 查看存储配额与用量 |

## 资源

| URI | 需要 Token |
|-----|:----------:|
| `picfast://user/profile` | 是 |
| `picfast://images/{key}` | 是 |

## 游客模式

不设置 `PICFAST_API_TOKEN` 时，仅 `upload_image` 可用。上传的图片以匿名方式存储，不绑定账户。游客上传需在 PicFast 实例中开启（管理员设置 → 允许游客上传）。若游客上传被禁用，将收到错误提示，此时需设置 `PICFAST_API_TOKEN` 进行认证。

## 许可证

MIT