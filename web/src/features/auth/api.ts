import { apiFetch, apiPost } from "@/shared/api/client";

import type { AuthSession } from "./types";

interface AuthResponse {
  data: AuthSession;
}

export async function getAuthSession(): Promise<AuthSession> {
  const response = await apiFetch<AuthResponse>("/api/v1/auth/session");
  return response.data;
}

export async function login(token: string): Promise<AuthSession> {
  const response = await apiPost<AuthResponse>("/api/v1/auth/login", { token });
  return response.data;
}

export async function logout(): Promise<void> {
  await apiPost<void>("/api/v1/auth/logout");
}
