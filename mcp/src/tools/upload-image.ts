import { z } from "zod";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";
import { existsSync } from "fs";
import { resolve, extname, basename } from "path";

const IMAGE_EXTENSIONS = new Set([
  ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg", ".ico", ".tiff", ".tif", ".avif",
]);

export function registerUploadImageTool(server: McpServer, client: ApiClient) {
  server.tool(
    "upload_image",
    "Upload an image from a local file path to PicFast. Returns the image key, URL, and formatted links. Works without PICFAST_API_TOKEN for guest uploads (requires guest upload to be enabled on the PicFast instance).",
    {
      file_path: z.string().describe("Absolute or relative path to the image file on the local filesystem."),
      filename: z.string().optional().describe("Override filename with extension. Defaults to the basename of file_path."),
      album_id: z.number().int().optional().describe("Target album ID. Requires PICFAST_API_TOKEN."),
      permission: z.number().int().optional().describe("Image visibility: 0=private, 1=public. Requires PICFAST_API_TOKEN."),
    },
    async (args) => {
      const absPath = resolve(args.file_path);

      if (!existsSync(absPath)) {
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "FILE_NOT_FOUND", message: `File does not exist: ${absPath}` } }) }],
          isError: true,
        };
      }

      const ext = extname(absPath).toLowerCase();
      if (!IMAGE_EXTENSIONS.has(ext)) {
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "INVALID_FILE_TYPE", message: `Unsupported image extension: ${ext}` } }) }],
          isError: true,
        };
      }

      try {
        const result = await client.uploadImage(absPath, {
          filename: args.filename,
          album_id: args.album_id,
          permission: args.permission,
        });

        const response = {
          key: result.key,
          url: result.links.url,
          markdown: result.links.markdown,
          html: result.links.html,
          bbcode: result.links.bbcode,
          mimetype: result.mimetype,
          original_size: result.size_bytes,
          width: result.width,
          height: result.height,
        };

        return {
          content: [{ type: "text" as const, text: JSON.stringify(response) }],
        };
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "UPLOAD_FAILED", message } }) }],
          isError: true,
        };
      }
    }
  );
}