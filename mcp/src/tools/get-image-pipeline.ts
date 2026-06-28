import { z } from "zod";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerGetImagePipelineTool(server: McpServer, client: ApiClient) {
  server.tool(
    "get_image_pipeline",
    "Get the processing pipeline status for an image. Returns upload, processing, thumbnail, and moderation stages. Requires PICFAST_API_TOKEN.",
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
        const pipeline = await client.getPipelineStatus(args.key);
        return {
          content: [{ type: "text" as const, text: JSON.stringify(pipeline) }],
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
