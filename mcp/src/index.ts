#!/usr/bin/env node

import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { loadConfig } from "./config.js";
import { ApiClient } from "./api-client.js";
import { registerUploadImageTool } from "./tools/upload-image.js";
import { registerListImagesTool } from "./tools/list-images.js";
import { registerGetImageTool } from "./tools/get-image.js";
import { registerDeleteImageTool } from "./tools/delete-image.js";
import { registerUsageStatsTool } from "./tools/usage-stats.js";
import { registerUserProfileResource } from "./resources/user-profile.js";
import { registerImageDetailResource } from "./resources/image-detail.js";

async function main() {
  const config = loadConfig();
  const client = new ApiClient(config);

  const server = new McpServer({
    name: "picfast",
    version: "0.1.0",
  });

  registerUploadImageTool(server, client);
  registerListImagesTool(server, client);
  registerGetImageTool(server, client);
  registerDeleteImageTool(server, client);
  registerUsageStatsTool(server, client);

  registerUserProfileResource(server, client);
  registerImageDetailResource(server, client);

  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((err) => {
  console.error("Fatal error starting PicFast MCP server:", err);
  process.exit(1);
});