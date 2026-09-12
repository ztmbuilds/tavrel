import { apiFetch } from "@/shared/api/client";
import type { Log, LogLevel } from "@/shared/types/entities/log";
import type { Span, SpanKind, SpanStatus } from "@/shared/types/entities/span";
import type { Trace, TraceStatus } from "@/shared/types/entities/trace";

import type {
  SpanEventView,
  SpanNode,
  TraceDetail,
  TracesListParams,
} from "./types";

export const TRACES_PAGE_SIZE = 20;
export const SPAN_LOGS_LIMIT = 100;

const WARN_SEVERITY_MIN = 13;
const WARN_SEVERITY_MAX = 16;
const ERROR_SEVERITY_MIN = 17;

const SPAN_KIND_BY_INDEX: readonly SpanKind[] = [
  "unspecified",
  "internal",
  "server",
  "client",
  "producer",
  "consumer",
];

function toSpanKind(kind: number): SpanKind {
  return SPAN_KIND_BY_INDEX[kind] ?? "unspecified";
}

function toSpanStatus(statusCode: number): SpanStatus {
  if (statusCode === 1) return "ok";
  if (statusCode === 2) return "error";
  return "unset";
}

export function isAsyncSpan(kind: SpanKind): boolean {
  return kind === "producer" || kind === "consumer";
}

function toTraceStatus(error: boolean): TraceStatus {
  return error ? "error" : "ok";
}

interface TraceDto {
  traceId: string;
  rootService: string;
  rootName: string;
  environment: string;
  startTime: string;
  endTime: string;
  spanCount: number;
  error: boolean;
}

interface TracesResponse {
  message: string;
  data: TraceDto[];
  meta: { nextCursor: string | null; total?: number | null };
}

interface ServicesResponse {
  message: string;
  data: string[];
  meta: { nextCursor: string | null };
}

interface EnvironmentsResponse {
  message: string;
  data: string[];
  meta: { nextCursor: string | null };
}

function toTrace(dto: TraceDto): Trace {
  return {
    traceId: dto.traceId,
    rootService: dto.rootService,
    rootName: dto.rootName,
    environment: dto.environment,
    startTime: dto.startTime,
    endTime: dto.endTime,
    spanCount: dto.spanCount,
    status: toTraceStatus(dto.error),
  };
}

export interface TracesPage {
  traces: Trace[];
  nextCursor: string | null;
  total: number | null;
}

export async function getTraces(params: TracesListParams): Promise<TracesPage> {
  const query: Record<string, string | number | undefined> = {
    limit: TRACES_PAGE_SIZE,
  };
  if (params.search !== "") query.search = params.search;
  if (params.status !== "all") query.status = params.status;
  if (params.service !== "" && params.service !== "all")
    query.service = params.service;
  if (params.environment !== "" && params.environment !== "all")
    query.environment = params.environment;
  if (params.order !== "desc") query.order = params.order;
  if (params.start !== undefined) query.start = params.start;
  if (params.end !== undefined) query.end = params.end;
  if (params.cursor !== undefined) query.cursor = params.cursor;

  const response = await apiFetch<TracesResponse>("/api/v1/traces", { query });
  return {
    traces: response.data.map(toTrace),
    nextCursor: response.meta.nextCursor,
    total: response.meta.total ?? null,
  };
}

export async function getServices(): Promise<string[]> {
  const response = await apiFetch<ServicesResponse>("/api/v1/services");
  return response.data;
}

export async function getEnvironments(): Promise<string[]> {
  const response =
    await apiFetch<EnvironmentsResponse>("/api/v1/environments");
  return response.data;
}

interface SpanDto {
  spanId: string;
  traceId: string;
  parentSpanId: string;
  service: string;
  name: string;
  kind: number;
  statusCode: number;
  statusMessage: string;
  startTime: string;
  endTime: string;
  attributes: unknown;
  events: unknown;
}

interface SpanNodeDto {
  span: SpanDto;
  children: SpanNodeDto[];
}

interface TraceDetailData {
  trace: TraceDto;
  spans: SpanNodeDto[];
}

interface TraceDetailResponse {
  message: string;
  data: TraceDetailData;
  meta: { nextCursor: string | null };
}

function toSpan(dto: SpanDto): Span {
  return {
    spanId: dto.spanId,
    traceId: dto.traceId,
    parentSpanId: dto.parentSpanId,
    service: dto.service,
    name: dto.name,
    kind: toSpanKind(dto.kind),
    status: toSpanStatus(dto.statusCode),
    statusMessage: dto.statusMessage,
    startTime: dto.startTime,
    endTime: dto.endTime,
    attributes: dto.attributes,
    events: dto.events,
  };
}

function toSpanNode(dto: SpanNodeDto): SpanNode {
  return {
    span: toSpan(dto.span),
    children: dto.children.map(toSpanNode),
  };
}

export async function getTraceDetail(traceId: string): Promise<TraceDetail> {
  const response = await apiFetch<TraceDetailResponse>(
    `/api/v1/traces/${encodeURIComponent(traceId)}`,
  );
  return {
    trace: toTrace(response.data.trace),
    spans: response.data.spans.map(toSpanNode),
  };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

interface RawSpanEvent {
  time: string;
  name: string;
  attributes: unknown;
}

function toRawSpanEvent(value: unknown): RawSpanEvent | null {
  if (!isRecord(value)) return null;
  if (typeof value["time"] !== "string") return null;
  if (typeof value["name"] !== "string") return null;
  return {
    time: value["time"],
    name: value["name"],
    attributes: value["attributes"],
  };
}

function toEventDetail(attributes: unknown): string | undefined {
  if (!isRecord(attributes)) return undefined;
  const message = attributes["exception.message"];
  const type = attributes["exception.type"];
  const parts: string[] = [];
  if (typeof type === "string" && type !== "") parts.push(type);
  if (typeof message === "string" && message !== "") parts.push(message);
  if (parts.length > 0) return parts.join(": ");
  const keys = Object.keys(attributes);
  if (keys.length === 0) return undefined;
  try {
    return JSON.stringify(attributes);
  } catch {
    return undefined;
  }
}

export function toSpanEventViews(
  events: unknown,
  spanStartTime: string,
): SpanEventView[] {
  if (!Array.isArray(events)) return [];
  const spanStartMs = Date.parse(spanStartTime);
  const views: SpanEventView[] = [];
  for (const item of events) {
    const raw = toRawSpanEvent(item);
    if (raw === null) continue;
    const offsetMs = Math.max(Date.parse(raw.time) - spanStartMs, 0);
    const view: SpanEventView = {
      offsetMs: Number.isNaN(offsetMs) ? 0 : offsetMs,
      name: raw.name,
      type: raw.name.toLowerCase() === "exception" ? "error" : "info",
    };
    const detail = toEventDetail(raw.attributes);
    if (detail !== undefined) view.detail = detail;
    views.push(view);
  }
  return views.sort((a, b) => a.offsetMs - b.offsetMs);
}

function toLogLevel(severityNumber: number): LogLevel {
  if (severityNumber >= ERROR_SEVERITY_MIN) return "error";
  if (severityNumber >= WARN_SEVERITY_MIN && severityNumber <= WARN_SEVERITY_MAX)
    return "warn";
  return "info";
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
    host:
      lookupHost(dto.attributes) ?? lookupHost(dto.resourceAttributes) ?? "",
    attributes: dto.attributes,
    resourceAttributes: dto.resourceAttributes,
  };
}

export async function getSpanLogs(
  traceId: string,
  spanId: string,
  start: string,
  end: string,
): Promise<Log[]> {
  const response = await apiFetch<LogsResponse>("/api/v1/logs", {
    query: { limit: SPAN_LOGS_LIMIT, traceId, spanId, start, end },
  });
  return response.data.map(toLog);
}
