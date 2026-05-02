import { z } from "zod";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerListImagesTool(server: McpServer, client: ApiClient) {
  server.tool(
    "list_images",
    "List your uploaded images with pagination. Requires PICFAST_API_TOKEN.",
    {
      page: z.number().int().optional().describe("Page number, starting from 1."),
      page_size: z.number().int().optional().describe("Page size. Default 20, max 100."),
    },
    async (args) => {
      if (!client.authenticated) {
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "UNAUTHORIZED", message: "PICFAST_API_TOKEN is required for this operation." } }) }],
          isError: true,
        };
      }

      try {
        const result = await client.listImages(args.page, args.page_size);
        return {
          content: [{ type: "text" as const, text: JSON.stringify(result) }],
        };
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "INTERNAL_ERROR", message } }) }],
          isError: true,
        };
      }
    }
  );
}