import type { AxiosInstance, AxiosRequestConfig } from "axios";
import type { z } from "zod";

import axios from "axios";

const HTTP_STATUS_NOT_FOUND = 404;
const HTTP_STATUS_BAD_REQUEST = 400;
const HTTP_STATUS_SCHEMA_MISMATCH = 500;
const DEFAULT_TIMEOUT_MS = 30_000;

export class ApiError extends Error {
  readonly status: number;
  readonly path: string;

  constructor(status: number, path: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.path = path;
  }

  isNotFound(): boolean {
    return this.status === HTTP_STATUS_NOT_FOUND;
  }

  isBadRequest(): boolean {
    return this.status === HTTP_STATUS_BAD_REQUEST;
  }
}

interface ApiFetchOptions {
  query?: Record<string, string | number | undefined>;
  config?: Omit<AxiosRequestConfig, "url" | "method" | "params" | "baseURL">;
}

const http: AxiosInstance = axios.create({
  headers: { Accept: "application/json" },
  timeout: DEFAULT_TIMEOUT_MS,
});

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function extractMessage(payload: unknown, fallback: string): string {
  if (isRecord(payload) && typeof payload["error"] === "string")
    return payload["error"];
  if (isRecord(payload) && typeof payload["message"] === "string")
    return payload["message"];
  return fallback;
}

function toApiError(error: unknown, url: string): ApiError {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status ?? 0;
    if (status === 0) {
      return new ApiError(0, url, error.message || "Network request failed");
    }
    const data: unknown = error.response?.data;
    return new ApiError(
      status,
      url,
      extractMessage(data, `Request failed with status ${status}`),
    );
  }
  return new ApiError(
    0,
    url,
    error instanceof Error ? error.message : "Unknown request failure",
  );
}

export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  try {
    const response = await http.get<T>(path, {
      params: options.query,
      ...options.config,
    });
    return response.data;
  } catch (error: unknown) {
    throw toApiError(error, path);
  }
}

// Fetch + validate at the boundary with zod. Parses unknown JSON so a shape
// drift fails loudly instead of flowing mistyped data into entities.
export async function apiFetchValidated<T>(
  path: string,
  schema: z.ZodType<T>,
  options: ApiFetchOptions = {},
): Promise<T> {
  const payload: unknown = await apiFetch<unknown>(path, options);
  const parsed = schema.safeParse(payload);
  if (!parsed.success) {
    throw new ApiError(
      HTTP_STATUS_SCHEMA_MISMATCH,
      path,
      `Invalid API response shape: ${parsed.error.issues[0]?.message ?? "unknown"}`,
    );
  }
  return parsed.data;
}
