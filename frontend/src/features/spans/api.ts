import { apiFetch } from "@/shared/api/client";
import { toSpan, type SpanDto } from "@/shared/api/otel";
import type { Span } from "@/shared/types/entities/span";

import type { SpansListParams } from "./types";

export const SPANS_PAGE_SIZE = 25;

interface SpansResponse {
  message: string;
  data: SpanDto[];
  meta: { nextCursor: string | null };
}

export interface SpansPage {
  spans: Span[];
  nextCursor: string | null;
}

export async function getSpans(params: SpansListParams): Promise<SpansPage> {
  const query: Record<string, string | number | string[] | undefined> = {
    limit: SPANS_PAGE_SIZE,
  };
  if (params.search !== "") query.search = params.search;
  if (params.status !== "all") query.status = params.status;
  if (params.service.length > 0) query.service = params.service;
  if (params.environment.length > 0) query.environment = params.environment;
  if (params.start !== undefined) query.start = params.start;
  if (params.end !== undefined) query.end = params.end;
  if (params.cursor !== undefined) query.cursor = params.cursor;

  const response = await apiFetch<SpansResponse>("/api/v1/spans", { query });
  return {
    spans: response.data.map(toSpan),
    nextCursor: response.meta.nextCursor,
  };
}
