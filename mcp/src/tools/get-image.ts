import { z } from "zod";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerGetImageTool(server: McpServer, client: ApiClient) {
  server.tool(
    "get_image",
    "Get details and shareable links for a specific image by its key. Requires PICFAST_API_TOKEN.",
    {
      key: z.string().describe("Image key."),
    },
    async (args) => {
      if (!client.authenticated) {
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "UNAUTHORIZED", message: "PICFAST_API_TOKEN is required for this operation." } }) }],
          isError: true,
        };
      }

      try {
        const img = await client.getImage(args.key);
        const response = {
          key: img.key,
          origin_name: img.origin_name,
          size_bytes: img.size_bytes,
          mimetype: img.mimetype,
          width: img.width,
          height: img.height,
          permission: img.permission,
          url: img.links.url,
          thumbnail: img.links.thumbnail_url,
          markdown: img.links.markdown,
          html: img.links.html,
          bbcode: img.links.bbcode,
          created_at: img.created_at,
        };
        return {
          content: [{ type: "text" as const, text: JSON.stringify(response) }],
        };
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "IMAGE_NOT_FOUND", message } }) }],
          isError: true,
        };
      }
    }
  );
}