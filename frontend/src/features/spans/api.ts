import { apiFetch } from "@/shared/api/client";
import type { Span, SpanKind, SpanStatus } from "@/shared/types/entities/span";

import type { SpansListParams } from "./types";

export const SPANS_PAGE_SIZE = 25;

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

interface SpansResponse {
  message: string;
  data: SpanDto[];
  meta: { nextCursor: string | null; total?: number | null };
}

interface ServicesResponse {
  message: string;
  data: string[];
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

export interface SpansPage {
  spans: Span[];
  nextCursor: string | null;
  total: number | null;
}

export async function getSpans(params: SpansListParams): Promise<SpansPage> {
  const query: Record<string, string | number | undefined> = {
    limit: SPANS_PAGE_SIZE,
  };
  if (params.search !== "") query.search = params.search;
  if (params.status !== "all") query.status = params.status;
  if (params.service !== "" && params.service !== "all")
    query.service = params.service;
  if (params.start !== undefined) query.start = params.start;
  if (params.end !== undefined) query.end = params.end;
  if (params.cursor !== undefined) query.cursor = params.cursor;

  const response = await apiFetch<SpansResponse>("/api/v1/spans", { query });
  return {
    spans: response.data.map(toSpan),
    nextCursor: response.meta.nextCursor,
    total: response.meta.total ?? null,
  };
}

export async function getServices(): Promise<string[]> {
  const response = await apiFetch<ServicesResponse>("/api/v1/services");
  return response.data;
}
