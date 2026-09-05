const HTTP_STATUS_NOT_FOUND = 404;
const HTTP_STATUS_BAD_REQUEST = 400;
const DEFAULT_TIMEOUT_MS = 30_000;
const HTTP_STATUS_NO_CONTENT = 204;

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
  config?: Omit<RequestInit, "method" | "body">;
}

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

function buildUrl(
  path: string,
  query: ApiFetchOptions["query"],
): string {
  if (!query) return path;
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value !== undefined) params.append(key, String(value));
  }
  const suffix = params.toString();
  if (!suffix) return path;
  return path + (path.includes("?") ? "&" : "?") + suffix;
}

function withTimeout(userSignal: AbortSignal | null | undefined): {
  signal: AbortSignal;
  cleanup: () => void;
} {
  const controller = new AbortController();
  const timeoutId = setTimeout(() => {
    controller.abort(
      new DOMException(
        `Request timed out after ${DEFAULT_TIMEOUT_MS}ms`,
        "TimeoutError",
      ),
    );
  }, DEFAULT_TIMEOUT_MS);
  const onUserAbort = (): void => {
    controller.abort(userSignal?.reason);
  };
  if (userSignal) {
    if (userSignal.aborted) {
      onUserAbort();
    } else {
      userSignal.addEventListener("abort", onUserAbort, { once: true });
    }
  }
  const cleanup = (): void => {
    clearTimeout(timeoutId);
    userSignal?.removeEventListener("abort", onUserAbort);
  };
  return { signal: controller.signal, cleanup };
}

async function parseJsonSafe(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return undefined;
  }
}

export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const url = buildUrl(path, options.query);
  const { signal: userSignal, headers: userHeaders, ...rest } =
    options.config ?? {};
  const headers = new Headers(userHeaders);
  if (!headers.has("Accept")) headers.set("Accept", "application/json");
  const { signal, cleanup } = withTimeout(userSignal);

  try {
    const response = await fetch(url, {
      ...rest,
      method: "GET",
      headers,
      signal,
    });
    if (!response.ok) {
      const payload = await parseJsonSafe(response).catch(
        () => undefined as unknown,
      );
      throw new ApiError(
        response.status,
        path,
        extractMessage(
          payload,
          `Request failed with status ${response.status}`,
        ),
      );
    }
    if (response.status === HTTP_STATUS_NO_CONTENT) return undefined as T;
    const data = await parseJsonSafe(response);
    return data as T;
  } catch (error: unknown) {
    if (error instanceof ApiError) throw error;
    throw new ApiError(
      0,
      path,
      error instanceof Error ? error.message : "Network request failed",
    );
  } finally {
    cleanup();
  }
}
