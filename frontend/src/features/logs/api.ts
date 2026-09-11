import { apiFetch } from "@/shared/api/client";
import type { Log, LogLevel } from "@/shared/types/entities/log";

import type { LogsListParams } from "./types";

export const LOGS_PAGE_SIZE = 30;

const WARN_SEVERITY_MIN = 13;
const ERROR_SEVERITY_MIN = 17;
const WARN_SEVERITY_MAX = 16;

function toLogLevel(severityNumber: number): LogLevel {
  if (severityNumber >= ERROR_SEVERITY_MIN) return "error";
  if (
    severityNumber >= WARN_SEVERITY_MIN &&
    severityNumber <= WARN_SEVERITY_MAX
  )
    return "warn";
  return "info";
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function toLogMessage(body: unknown): string {
  if (typeof body === "string") return body;
  if (isRecord(body)) {
    const stringValue = body["stringValue"];
    if (typeof stringValue === "string") return stringValue;
    const message = body["message"];
    if (typeof message === "string") return message;
  }
  try {
    return JSON.stringify(body);
  } catch {
    return String(body);
  }
}

const HOST_KEYS: readonly string[] = ["host.name", "host", "hostname"];

function lookupHost(value: unknown): string | undefined {
  if (!isRecord(value)) return undefined;
  for (const key of HOST_KEYS) {
    const candidate = value[key];
    if (typeof candidate === "string" && candidate !== "") return candidate;
  }
  return undefined;
}

function toLogHost(
  attributes: unknown,
  resourceAttributes: unknown,
): string {
  return (
    lookupHost(attributes) ?? lookupHost(resourceAttributes) ?? ""
  );
}

interface LogDto {
  id: number;
  traceId: string;
  spanId: string;
  timestamp: string;
  severityText: string;
  severityNumber: number;
  body: unknown;
  service: string;
  environment: string;
  attributes: unknown;
  resourceAttributes: unknown;
}

interface LogsResponse {
  message: string;
  data: LogDto[];
  meta: { nextCursor: string | null; total?: number | null };
}

interface ServicesResponse {
  message: string;
  data: string[];
  meta: { nextCursor: string | null };
}

function toLog(dto: LogDto): Log {
  return {
    id: dto.id,
    traceId: dto.traceId,
    spanId: dto.spanId,
    timestamp: dto.timestamp,
    service: dto.service,
    environment: dto.environment,
    severityText: dto.severityText,
    severityNumber: dto.severityNumber,
    level: toLogLevel(dto.severityNumber),
    message: toLogMessage(dto.body),
    host: toLogHost(dto.attributes, dto.resourceAttributes),
    attributes: dto.attributes,
    resourceAttributes: dto.resourceAttributes,
  };
}

export interface LogsPage {
  logs: Log[];
  nextCursor: string | null;
  total: number | null;
}

export async function getLogs(params: LogsListParams): Promise<LogsPage> {
  const query: Record<string, string | number | undefined> = {
    limit: LOGS_PAGE_SIZE,
  };
  if (params.search !== "") query.search = params.search;
  if (params.level !== "all") query.level = params.level;
  if (params.service !== "" && params.service !== "all")
    query.service = params.service;
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
  const response = await apiFetch<ServicesResponse>("/api/v1/services");
  return response.data;
}
