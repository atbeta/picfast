// src/config.ts
// Configuration access wrapper for the PicFast extension.

import * as vscode from "vscode";

export const CONFIG_SECTION = "picfast";
export const UPLOAD_PATH = "/api/v1/flat/upload";

export type InsertFormat = "url" | "markdown" | "html" | "bbcode";

export interface PicFastConfig {
  baseUrl: string;
  apiToken: string;
  defaultFormat: InsertFormat;
  timeoutMs: number;
  showStatusBar: boolean;
}

export function getConfig(): PicFastConfig {
  const cfg = vscode.workspace.getConfiguration(CONFIG_SECTION);
  const rawBase = (cfg.get<string>("baseUrl") ?? "").trim();
  const baseUrl = rawBase.replace(/\/+$/, "");
  return {
    baseUrl,
    apiToken: (cfg.get<string>("apiToken") ?? "").trim(),
    defaultFormat: cfg.get<InsertFormat>("defaultFormat") ?? "markdown",
    timeoutMs: cfg.get<number>("timeoutMs") ?? 30000,
    showStatusBar: cfg.get<boolean>("showStatusBar") ?? true,
  };
}

export function getUploadUrl(baseUrl: string): string {
  return `${baseUrl}${UPLOAD_PATH}`;
}
