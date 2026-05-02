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

    if (!res.ok) {
      const body = await res.text();
      throw new Error(`upload failed (${res.status}): ${body}`);
    }

    return (await res.json()) as UploadResult;
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

    if (!res.ok) {
      const body = await res.text();
      throw new Error(`list images failed (${res.status}): ${body}`);
    }

    return (await res.json()) as ImageListResult;
  }

  async getImage(key: string): Promise<UploadResult> {
    this.requireAuth();
    const res = await fetch(this.url(`/images/${key}`), {
      headers: this.authHeaders(),
    });

    if (!res.ok) {
      const body = await res.text();
      throw new Error(`get image failed (${res.status}): ${body}`);
    }

    return (await res.json()) as UploadResult;
  }

  async deleteImage(key: string): Promise<void> {
    this.requireAuth();
    const res = await fetch(this.url(`/images/${key}`), {
      method: "DELETE",
      headers: this.authHeaders(),
    });

    if (!res.ok) {
      const body = await res.text();
      throw new Error(`delete image failed (${res.status}): ${body}`);
    }
  }

  async getUserProfile(): Promise<UserProfile> {
    this.requireAuth();
    const res = await fetch(this.url("/users/me"), {
      headers: this.authHeaders(),
    });

    if (!res.ok) {
      const body = await res.text();
      throw new Error(`get profile failed (${res.status}): ${body}`);
    }

    return (await res.json()) as UserProfile;
  }
}