# picfast

This **npm package name is reserved** for the [PicFast](https://github.com/atbeta/picfast) project (image hosting /图床).

It does **not** ship a Node.js runtime or CLI here. The application is primarily the **Go** server in this repository (`cmd/picfast`).

## Integrations

| Use case | Package / path |
|----------|----------------|
| **MCP (local, Cursor / Claude / etc.)** | [`@picfast/mcp-server`](https://www.npmjs.com/package/@picfast/mcp-server) — uploads via `file_path` + REST multipart |
| **Server / binary** | Build from source: `cmd/picfast` |

## Why this package exists

The unscoped name `picfast` on npm is held so it cannot be squatted by unrelated projects. No functional code is published under this package.

---

# picfast（npm 占位说明）

本 **无 scope** 的 npm 包名保留给 [PicFast](https://github.com/atbeta/picfast) 官方项目使用；**此处不提供可运行的 Node 模块或 CLI**，服务端实现见仓库中的 **Go**（`cmd/picfast`）。

## 集成方式

| 场景 | 说明 |
|------|------|
| **本地 MCP** | 请使用 [`@picfast/mcp-server`](https://www.npmjs.com/package/@picfast/mcp-server)，通过本地路径读文件并以 multipart 上传 |
| **自建服务** | 从本仓库编译 `cmd/picfast` |

发布本占位包是为了避免第三方占用 `picfast` 这一 npm 包名；不包含业务代码。
