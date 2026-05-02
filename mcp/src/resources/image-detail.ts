import { McpServer, ResourceTemplate } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerImageDetailResource(server: McpServer, client: ApiClient) {
  server.registerResource(
    "image_detail",
    new ResourceTemplate("picfast://images/{key}", {
      list: async () => {
        if (!client.authenticated) {
          return { resources: [] };
        }
        try {
          const result = await client.listImages();
          return {
            resources: result.items.map((img) => ({
              uri: `picfast://images/${img.key}`,
              name: img.origin_name,
            })),
          };
        } catch {
          return { resources: [] };
        }
      },
    }),
    {
      description: "Get details for a specific image by key. Requires PICFAST_API_TOKEN.",
      mimeType: "application/json",
    },
    async (uri, { key }) => {
      if (!client.authenticated) {
        throw new Error("PICFAST_API_TOKEN is required to access this resource.");
      }

      try {
        const img = await client.getImage(key as string);
        return {
          contents: [
            {
              uri: uri.href,
              mimeType: "application/json",
              text: JSON.stringify({
                key: img.key,
                origin_name: img.origin_name,
                size_bytes: img.size_bytes,
                width: img.width,
                height: img.height,
                url: img.links.url,
                created_at: img.created_at,
              }),
            },
          ],
        };
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        throw new Error(`Failed to fetch image: ${message}`);
      }
    }
  );
}