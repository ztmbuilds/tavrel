import { apiFetch } from "@/shared/api/client";
import {
  toLog,
  type LogDto,
  type StringListResponse,
} from "@/shared/api/otel";
import type { Log } from "@/shared/types/entities/log";

import type { LogsListParams } from "./types";

export const LOGS_PAGE_SIZE = 30;

interface LogsResponse {
  message: string;
  data: LogDto[];
  meta: { nextCursor: string | null; total?: number | null };
}

export interface LogsPage {
  logs: Log[];
  nextCursor: string | null;
  total: number | null;
}

export async function getLogs(params: LogsListParams): Promise<LogsPage> {
  const query: Record<string, string | number | string[] | undefined> = {
    limit: LOGS_PAGE_SIZE,
  };
  if (params.search !== "") query.search = params.search;
  if (params.level !== "all") query.level = params.level;
  if (params.service.length > 0) query.service = params.service;
  if (params.traceId !== "") query.traceId = params.traceId;
  if (params.spanId !== "") query.spanId = params.spanId;
  if (params.start !== undefined) query.start = params.start;
  if (params.end !== undefined) query.end = params.end;
  if (params.cursor !== undefined) query.cursor = params.cursor;

  const response = await apiFetch<LogsResponse>("/api/v1/logs", { query });
  return {
    logs: response.data.map(toLog),
    nextCursor: response.meta.nextCursor,
    total: response.meta.total ?? null,
  };
}

export async function getServices(): Promise<string[]> {
  const response = await apiFetch<StringListResponse>("/api/v1/services");
  return response.data;
}
