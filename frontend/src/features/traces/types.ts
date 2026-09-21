import type { SharedTime } from "@/shared/composables/useGlobalTimeRange";
import type { Span, SpanStatus } from "@/shared/types/entities/span";
import type { Trace } from "@/shared/types/entities/trace";

export type TraceStatusFilter = "all" | "error" | "ok";

export type TraceOrder = "asc" | "desc";

export interface TracesFilters extends SharedTime {
  search: string;
  status: TraceStatusFilter;
  service: string[];
  environment: string[];
  order: TraceOrder;
}

export interface TracesListParams extends TracesFilters {
  cursor?: string;
}

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
  service: string;
  headline: string;
  tail: string;
}
