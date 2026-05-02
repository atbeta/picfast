import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerUsageStatsTool(server: McpServer, client: ApiClient) {
  server.tool(
    "get_usage_stats",
    "Get your storage usage statistics (used bytes, total capacity, image count). Requires PICFAST_API_TOKEN.",
    {},
    async () => {
      if (!client.authenticated) {
        return {
          content: [{ type: "text" as const, text: JSON.stringify({ error: { code: "UNAUTHORIZED", message: "PICFAST_API_TOKEN is required for this operation." } }) }],
          isError: true,
        };
      }

      try {
        const profile = await client.getUserProfile();
        const response = {
          user_id: profile.id,
          name: profile.name,
          capacity_bytes: profile.capacity_bytes,
          used_bytes: profile.used_bytes,
          image_num: profile.image_num,
          album_num: profile.album_num,
        };
        return {
          content: [{ type: "text" as const, text: JSON.stringify(response) }],
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