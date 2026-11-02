import { apiFetch } from "@/shared/api/client";
import {
  toLog,
  toSpan,
  toSpanLink,
  toTrace,
  toTraceResources,
  type LogDto,
  type SpanDto,
  type SpanLinkDto,
  type TraceDto,
  type TraceResourceDto,
} from "@/shared/api/otel";
import type { Log } from "@/shared/types/entities/log";
import type { SpanKind } from "@/shared/types/entities/span";
import type { SpanLink } from "@/shared/types/entities/span-link";
import type { Trace } from "@/shared/types/entities/trace";

import type { SpanNode, TraceDetail, TracesListParams } from "./types";

export const TRACES_PAGE_SIZE = 20;
export const SPAN_LOGS_LIMIT = 100;

export function isAsyncSpan(kind: SpanKind): boolean {
  return kind === "producer" || kind === "consumer";
}

interface TracesResponse {
  message: string;
  data: TraceDto[];
  meta: { nextCursor: string | null };
}

export interface TracesPage {
  traces: Trace[];
  nextCursor: string | null;
}

export async function getTraces(params: TracesListParams): Promise<TracesPage> {
  const query: Record<string, string | number | string[] | undefined> = {
    limit: TRACES_PAGE_SIZE,
  };
  if (params.search !== "") query.search = params.search;
  if (params.status !== "all") query.status = params.status;
  if (params.service.length > 0) query.service = params.service;
  if (params.environment.length > 0) query.environment = params.environment;
  if (params.order !== "desc") query.order = params.order;
  if (params.minDurationMs !== undefined)
    query.minDurationMs = params.minDurationMs;
  if (params.maxDurationMs !== undefined)
    query.maxDurationMs = params.maxDurationMs;
  if (params.start !== undefined) query.start = params.start;
  if (params.end !== undefined) query.end = params.end;
  if (params.cursor !== undefined) query.cursor = params.cursor;

  const response = await apiFetch<TracesResponse>("/api/v1/traces", { query });
  return {
    traces: response.data.map(toTrace),
    nextCursor: response.meta.nextCursor,
  };
}

interface SpanNodeDto {
  span: SpanDto;
  children: SpanNodeDto[];
}

interface TraceDetailData {
  trace: TraceDto;
  spans: SpanNodeDto[];
  resources: TraceResourceDto[];
}

interface TraceDetailResponse {
  message: string;
  data: TraceDetailData;
  meta: { nextCursor: string | null };
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
    resources: toTraceResources(response.data.resources),
  };
}

interface LogsResponse {
  message: string;
  data: LogDto[];
  meta: { nextCursor: string | null };
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

interface SpanLinksResponse {
  message: string;
  data: SpanLinkDto[];
  meta: { nextCursor: string | null };
}

export async function getSpanLinks(
  traceId: string,
  spanId: string,
): Promise<SpanLink[]> {
  const response = await apiFetch<SpanLinksResponse>("/api/v1/spans/links", {
    query: { traceId, spanId },
  });
  return response.data.map(toSpanLink);
}
