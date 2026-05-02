import { z } from "zod";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerDeleteImageTool(server: McpServer, client: ApiClient) {
  server.tool(
    "delete_image",
    "Delete an image by its key. This action is irreversible. Requires PICFAST_API_TOKEN.",
    {
      key: z.string().describe("Image key to delete."),
    },
    async (args) => {
      if (!client.authenticated) {
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "UNAUTHORIZED", message: "PICFAST_API_TOKEN is required for this operation." } }) }],
          isError: true,
        };
      }

      try {
        await client.deleteImage(args.key);
        return {
          content: [{ type: "text" as const, text: "Image deleted successfully." }],
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