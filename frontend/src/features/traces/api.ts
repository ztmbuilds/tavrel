import { apiFetch } from "@/shared/api/client";
import type { Trace, TraceStatus } from "@/shared/types/entities/trace";

import type { TracesListParams } from "./types";

export const TRACES_PAGE_SIZE = 20;

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
