import { Config } from "./config.js";

export interface UploadResult {
  id: number;
  key: string;
  origin_name: string;
  size_bytes: number;
  mimetype: string;
  extension: string;
  width: number;
  height: number;
  md5: string;
  sha1: string;
  permission: number;
  album_id: number | null;
  moderation_status: string;
  links: {
    url: string;
    html: string;
    bbcode: string;
    markdown: string;
    thumbnail_url: string;
  };
  created_at: string;
}

export interface ImageListResult {
  items: ImageListItem[];
  total: number;
  page: number;
  size: number;
}

export interface ImageListItem {
  id: number;
  key: string;
  origin_name: string;
  size_bytes: number;
  mimetype: string;
  extension: string;
  width: number;
  height: number;
  permission: number;
  album_id: number | null;
  url: string;
  thumbnail_url: string;
  moderation_status: string;
  strategy_id: number | null;
  strategy_name: string;
  strategy_type: string;
  links: {
    url: string;
    html: string;
    bbcode: string;
    markdown: string;
    thumbnail_url: string;
  };
  created_at: string;
}

export interface UserProfile {
  id: number;
  email: string;
  name: string;
  role: string;
  status: number;
  capacity_bytes: number;
  used_bytes: number;
  image_num: number;
  album_num: number;
}

export interface PipelineStatus {
  upload: string;
  processing: string;
  thumbnail: string;
  moderation: string;
  updated_at: string;
}

export class ApiClient {
  private baseUrl: string;
  private token: string | undefined;

  constructor(config: Config) {
    this.baseUrl = config.PICFAST_BASE_URL.replace(/\/+$/, "");
    this.token = config.PICFAST_API_TOKEN;
  }

  get authenticated(): boolean {
    return !!this.token;
  }

  private requireAuth(): void {
    if (!this.token) {
      throw new Error("PICFAST_API_TOKEN is required for this operation. Set it in your MCP server configuration.");
    }
  }

  private authHeaders(): Record<string, string> {
    if (!this.token) return { Accept: "application/json" };
    return {
      Authorization: `Bearer ${this.token}`,
      Accept: "application/json",
    };
  }

  private url(path: string): string {
    return `${this.baseUrl}/api/v1${path}`;
  }

  /**
   * PicFast REST handlers wrap payloads as `{ status, message, data }`.
   * Unwrap `data` when present; otherwise return the body as-is.
   */
  private async handleResponse<T>(res: Response): Promise<T> {
    const text = await res.text();
    let parsed: unknown;
    try {
      parsed = text ? JSON.parse(text) : {};
    } catch {
      const preview = text.slice(0, 300);
      if (!res.ok) {
        throw new Error(`request failed (${res.status}): ${preview}`);
      }
      throw new Error(`invalid json (${res.status}): ${preview}`);
    }

    const obj = parsed as Record<string, unknown> | null;
    const isEnvelope =
      obj !== null &&
      typeof obj === "object" &&
      "status" in obj &&
      typeof (obj as { status?: unknown }).status === "boolean";

    if (isEnvelope) {
      const envelopeOk = (obj as { status: boolean }).status;
      const message =
        typeof (obj as { message?: unknown }).message === "string"
          ? (obj as { message: string }).message
          : "request failed";
      if (!envelopeOk || !res.ok) {
        throw new Error(`${res.status}: ${message}`);
      }
      if ("data" in obj) {
        return (obj as { data: T }).data;
      }
      throw new Error(`unexpected envelope (${res.status})`);
    }

    if (!res.ok) {
      throw new Error(`request failed (${res.status}): ${text.slice(0, 300)}`);
    }

    return parsed as T;
  }

  async uploadImage(
    filePath: string,
    options?: { filename?: string; album_id?: number; permission?: number; strategy_id?: number }
  ): Promise<UploadResult> {
    const { readFile } = await import("fs/promises");
    const { basename } = await import("path");

    const data = await readFile(filePath);
    const filename = options?.filename ?? basename(filePath);
    const file = new File([data], filename);
    const form = new FormData();
    form.set("file", file);

    if (options?.album_id != null) form.set("album_id", String(options.album_id));
    if (options?.permission != null) form.set("permission", String(options.permission));
    if (options?.strategy_id != null) form.set("strategy_id", String(options.strategy_id));

    const headers: Record<string, string> = {};
    if (this.token) {
      headers["Authorization"] = `Bearer ${this.token}`;
    }

    // Authenticated requests go to /images; guest uploads go to /upload.
    const uploadPath = this.token ? "/images" : "/upload";

    const res = await fetch(this.url(uploadPath), {
      method: "POST",
      headers,
      body: form,
    });

    try {
      return await this.handleResponse<UploadResult>(res);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`upload failed (${res.status}): ${message}`);
    }
  }

  async listImages(
    page?: number,
    pageSize?: number
  ): Promise<ImageListResult> {
    this.requireAuth();
    const params = new URLSearchParams();
    if (page) params.set("page", String(page));
    if (pageSize) params.set("page_size", String(pageSize));

    const qs = params.toString();
    const path = qs ? `/images?${qs}` : "/images";

    const res = await fetch(this.url(path), {
      headers: this.authHeaders(),
    });

    try {
      return await this.handleResponse<ImageListResult>(res);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`list images failed (${res.status}): ${message}`);
    }
  }

  async getImage(key: string): Promise<UploadResult> {
    this.requireAuth();
    const res = await fetch(this.url(`/images/${key}`), {
      headers: this.authHeaders(),
    });

    try {
      return await this.handleResponse<UploadResult>(res);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`get image failed (${res.status}): ${message}`);
    }
  }

  async deleteImage(key: string): Promise<void> {
    this.requireAuth();
    const res = await fetch(this.url(`/images/${key}`), {
      method: "DELETE",
      headers: this.authHeaders(),
    });

    try {
      await this.handleResponse<unknown>(res);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`delete image failed (${res.status}): ${message}`);
    }
  }

  async getUserProfile(): Promise<UserProfile> {
    this.requireAuth();
    const res = await fetch(this.url("/users/me"), {
      headers: this.authHeaders(),
    });

    try {
      return await this.handleResponse<UserProfile>(res);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`get profile failed (${res.status}): ${message}`);
    }
  }

  async getPipelineStatus(key: string): Promise<PipelineStatus> {
    this.requireAuth();
    const res = await fetch(this.url(`/images/${key}/pipeline`), {
      headers: this.authHeaders(),
    });

    try {
      return await this.handleResponse<PipelineStatus>(res);
    } catch (err) {
      const message = err instanceof Error ? err.message : String(err);
      throw new Error(`get pipeline failed (${res.status}): ${message}`);
    }
  }
}
