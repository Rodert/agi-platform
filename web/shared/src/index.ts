export type ApiEnvelope<T> = {
  code: number;
  message: string;
  data?: T;
};

export type AuthToken = {
  access_token: string;
  token_type: string;
  expires_in: number;
};

export type User = {
  id: number;
  email?: string;
  phone?: string;
  nickname: string;
  avatar_url: string;
  credits: number;
  status: string;
  created_at: string;
};

export type AdminUser = {
  id: number;
  username: string;
  nickname: string;
  role: string;
  status: string;
};

export type ImageModel = {
  id: number;
  code: string;
  display_name: string;
  description: string;
  cover_url: string;
  price_credits: number;
  supported_sizes: string[] | unknown;
  support_text_to_image: boolean;
  support_image_to_image: boolean;
  support_edit: boolean;
  max_images_per_request: number;
  auto_refund_on_failure: boolean;
  enabled: boolean;
  recommended: boolean;
  sort_order: number;
};

export type Provider = {
  id: number;
  code: string;
  name: string;
  type: string;
  base_url: string;
  enabled: boolean;
  timeout_seconds: number;
  retry_count: number;
  priority: number;
  daily_limit?: number;
  remark: string;
};

export type ProviderKey = {
  id: number;
  provider_id: number;
  name: string;
  status: string;
  weight: number;
  daily_limit?: number;
  daily_used: number;
  last_error: string;
};

export type ImageModelRoute = {
  id: number;
  model_id: number;
  provider_id: number;
  provider_key_id?: number;
  provider_model_name: string;
  enabled: boolean;
  priority: number;
  weight: number;
  extra_config?: unknown;
};

export type ImageAsset = {
  id: number;
  task_id: number;
  user_id: number;
  model_id: number;
  url: string;
  width?: number;
  height?: number;
  status: string;
};

export type ImageTask = {
  id: number;
  task_no: string;
  user_id: number;
  model_id: number;
  source: string;
  prompt: string;
  negative_prompt?: string;
  size: string;
  num_images: number;
  status: string;
  progress: number;
  credits_used: number;
  error_message: string;
  created_at: string;
};

export type WalletLog = {
  id: number;
  user_id: number;
  type: string;
  amount: number;
  balance_before: number;
  balance_after: number;
  related_type: string;
  related_id?: number;
  remark: string;
  operator_type: string;
  operator_id?: number;
  created_at: string;
};

export type VideoModel = {
  id: number;
  code: string;
  display_name: string;
  description: string;
  price_credits: number;
  supported_aspect_ratios: string[] | unknown;
  supported_seconds: number[] | unknown;
  enabled: boolean;
  recommended: boolean;
  sort_order: number;
};

export type VideoModelRoute = {
  id: number;
  model_id: number;
  provider_id: number;
  provider_key_id?: number;
  provider_model_name: string;
  enabled: boolean;
  priority: number;
  weight: number;
  extra_config?: unknown;
};

export type VideoTask = {
  id: number;
  task_no: string;
  user_id: number;
  model_id: number;
  source: string;
  prompt: string;
  seconds: number;
  aspect_ratio: string;
  images?: string[] | unknown;
  videos?: string[] | unknown;
  audios?: string[] | unknown;
  status: string;
  progress: number;
  credits_used: number;
  error_message: string;
  created_at: string;
};

export type VideoAsset = {
  id: number;
  task_id: number;
  user_id: number;
  model_id: number;
  url: string;
  mime_type: string;
  size_bytes?: number;
  duration_seconds?: number;
  status: string;
};

export type VideoTaskResult = {
  task: VideoTask;
  videos: VideoAsset[];
};

export type GenerateImageResult = {
  task: ImageTask;
  images: ImageAsset[];
};

export type DatabaseTable = {
  name: string;
  table_type: string;
  rows: number;
  comment: string;
};

export type DatabaseColumn = {
  name: string;
  type: string;
  nullable: boolean;
  primary_key: boolean;
  comment: string;
};

export type DatabaseTableData = {
  table: string;
  comment: string;
  columns: DatabaseColumn[];
  rows: Record<string, unknown>[];
  ddl: string;
  limit: number;
  offset: number;
  has_next: boolean;
};

export type ApiKey = {
  id: number;
  user_id: number;
  name: string;
  key_prefix: string;
  status: string;
  last_used_at?: string;
  created_at: string;
};

export type CreateApiKeyResult = {
  api_key: ApiKey;
  plain: string;
};

export class ApiError extends Error {
  status: number;
  code: number;

  constructor(status: number, code: number, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export type ApiClientOptions = {
  baseURL?: string;
  getToken?: () => string | null;
  onUnauthorized?: () => void;
};

export class ApiClient {
  private baseURL: string;
  private getToken?: () => string | null;
  private onUnauthorized?: () => void;

  constructor(options: ApiClientOptions = {}) {
    this.baseURL = options.baseURL ?? viteEnv("VITE_API_BASE_URL") ?? "http://127.0.0.1:8080";
    this.getToken = options.getToken;
    this.onUnauthorized = options.onUnauthorized;
  }

  get<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: "GET" });
  }

  post<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, { method: "POST", body: body === undefined ? undefined : JSON.stringify(body) });
  }

  put<T>(path: string, body?: unknown): Promise<T> {
    return this.request<T>(path, { method: "PUT", body: body === undefined ? undefined : JSON.stringify(body) });
  }

  delete<T>(path: string): Promise<T> {
    return this.request<T>(path, { method: "DELETE" });
  }

  upload<T>(path: string, formData: FormData): Promise<T> {
    return this.request<T>(path, { method: "POST", body: formData }, false);
  }

  private async request<T>(path: string, init: RequestInit, json = true): Promise<T> {
    const headers = new Headers(init.headers);
    if (json) {
      headers.set("Content-Type", "application/json");
    }
    const token = this.getToken?.();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }

    const response = await fetch(`${this.baseURL}${path}`, { ...init, headers });
    const payload = (await response.json().catch(() => ({ message: response.statusText }))) as ApiEnvelope<T>;
    if (!response.ok || payload.code !== 0) {
      if (response.status === 401) {
        this.onUnauthorized?.();
      }
      throw new ApiError(response.status, payload.code ?? response.status, payload.message || "request failed");
    }
    return payload.data as T;
  }
}

function viteEnv(key: string): string | undefined {
  const meta = import.meta as ImportMeta & { env?: Record<string, string | undefined> };
  return meta.env?.[key];
}

export function formatCredits(value: number | undefined): string {
  return `${value ?? 0} 积分`;
}

export function parseJSONList(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map(String);
  }
  if (typeof value === "string") {
    try {
      const parsed = JSON.parse(value);
      return Array.isArray(parsed) ? parsed.map(String) : [];
    } catch {
      return [];
    }
  }
  return [];
}
