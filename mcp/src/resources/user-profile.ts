import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { ApiClient } from "../api-client.js";

export function registerUserProfileResource(server: McpServer, client: ApiClient) {
  server.registerResource(
    "user_profile",
    "picfast://user/profile",
    {
      description: "Current user info and storage capacity. Requires PICFAST_API_TOKEN.",
      mimeType: "application/json",
    },
    async (uri) => {
      if (!client.authenticated) {
        throw new Error("PICFAST_API_TOKEN is required to access this resource.");
      }

      try {
        const profile = await client.getUserProfile();
        return {
          contents: [
            {
              uri: uri.href,
              mimeType: "application/json",
              text: JSON.stringify({
                id: profile.id,
                name: profile.name,
                email: profile.email,
                role: profile.role,
                capacity_bytes: profile.capacity_bytes,
                used_bytes: profile.used_bytes,
                image_num: profile.image_num,
                album_num: profile.album_num,
              }),
            },
          ],
        };
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        throw new Error(`Failed to fetch user profile: ${message}`);
      }
    }
  );
}