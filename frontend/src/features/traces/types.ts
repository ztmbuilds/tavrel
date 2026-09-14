import type { Span, SpanStatus } from "@/shared/types/entities/span";
import type { Trace } from "@/shared/types/entities/trace";

export type TraceStatusFilter = "all" | "error" | "ok";

export type TraceOrder = "asc" | "desc";

export interface TracesFilters {
  search: string;
  status: TraceStatusFilter;
  service: string;
  environment: string;
  order: TraceOrder;
  start?: string;
  end?: string;
}

export interface TracesListParams extends TracesFilters {
  cursor?: string;
}

export const DEFAULT_TRACES_FILTERS: TracesFilters = {
  search: "",
  status: "all",
  service: "all",
  environment: "all",
  order: "desc",
};

export interface SpanNode {
  span: Span;
  children: SpanNode[];
}

export interface TraceDetail {
  trace: Trace;
  spans: SpanNode[];
}

export type TraceDetailTab = "overview" | "attributes" | "events" | "logs";

export type SpanEventType = "error" | "info";

export interface SpanEventView {
  offsetMs: number;
  name: string;
  type: SpanEventType;
  detail?: string;
}

export interface SpanLogWindow {
  spanId: string;
  start: string;
  end: string;
}

export interface TraceIssue {
  spanId: string;
  name: string;
  service: string;
  durationMs: number;
  status: SpanStatus;
}

export interface TraceDiagnosis {
  kind: "error" | "ok";
  spanId: string | null;
  spanName: string;
  suffix: string;
  tail: string;
}
