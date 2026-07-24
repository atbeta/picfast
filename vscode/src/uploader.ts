// src/uploader.ts
// Upload helper: posts a multipart/form-data to PicFast's /api/v1/flat/upload
// and parses the response. Uses the global `fetch` (Node 18+ / VSCode 1.78+).

export interface UploadResponse {
  url: string;
  thumbnailUrl?: string;
  markdown?: string;
  html?: string;
  bbcode?: string;
  mimetype?: string;
  sizeBytes?: number;
  width?: number;
  height?: number;
}

export class UploadError extends Error {
  constructor(
    message: string,
    public readonly status?: number,
    public readonly body?: string,
  ) {
    super(message);
    this.name = "UploadError";
  }
}

export interface UploadOptions {
  baseUrl: string;
  apiToken?: string;
  timeoutMs: number;
  fileName: string;
  data: Uint8Array;
  signal?: AbortSignal;
}

export async function uploadImage(opts: UploadOptions): Promise<UploadResponse> {
  const url = `${opts.baseUrl.replace(/\/+$/, "")}/api/v1/flat/upload`;
  const form = new FormData();
  const blob = new Blob([opts.data], { type: "image/png" });
  form.append("file", blob, opts.fileName);

  const headers: Record<string, string> = {};
  if (opts.apiToken) {
    headers["Authorization"] = `Bearer ${opts.apiToken}`;
  }

  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), opts.timeoutMs);
  // Compose with caller-provided signal (if any).
  if (opts.signal) {
    if (opts.signal.aborted) {
      controller.abort();
    } else {
      opts.signal.addEventListener("abort", () => controller.abort());
    }
  }

  let resp: Response;
  try {
    resp = await fetch(url, {
      method: "POST",
      headers,
      body: form,
      signal: controller.signal,
    });
  } catch (err) {
    clearTimeout(timer);
    if (err instanceof Error && err.name === "AbortError") {
      throw new UploadError("Upload timed out");
    }
    throw new UploadError(
      `Network error: ${err instanceof Error ? err.message : String(err)}`,
    );
  } finally {
    clearTimeout(timer);
  }

  const text = await resp.text();
  if (!resp.ok) {
    throw new UploadError(
      `Upload failed (HTTP ${resp.status})`,
      resp.status,
      text,
    );
  }

  let json: unknown;
  try {
    json = JSON.parse(text);
  } catch {
    throw new UploadError(
      `Unexpected response (not JSON, HTTP ${resp.status})`,
      resp.status,
      text.slice(0, 200),
    );
  }

  if (!isUploadResponse(json)) {
    throw new UploadError(
      "Unexpected response shape: missing 'url' field",
      resp.status,
      text.slice(0, 200),
    );
  }

  return json;
}

function isUploadResponse(value: unknown): value is UploadResponse {
  if (typeof value !== "object" || value === null) {
    return false;
  }
  const url = (value as Record<string, unknown>)["url"];
  return typeof url === "string" && url.length > 0;
}

export function formatInsert(url: string, format: "url" | "markdown" | "html" | "bbcode", alt: string): string {
  switch (format) {
    case "url":
      return url;
    case "markdown":
      return `![${alt}](${url})`;
    case "html":
      return `<img src="${url}" alt="${alt}" />`;
    case "bbcode":
      return `[img]${url}[/img]`;
  }
}
