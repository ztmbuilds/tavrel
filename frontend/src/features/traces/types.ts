import type { Span, SpanStatus } from "@/shared/types/entities/span";
import type { Trace } from "@/shared/types/entities/trace";

const TraceStatusFilter = {
  All: "all",
  Error: "error",
  Ok: "ok",
} as const;
export type TraceStatusFilter =
  (typeof TraceStatusFilter)[keyof typeof TraceStatusFilter];

const TraceOrder = {
  Asc: "asc",
  Desc: "desc",
} as const;
export type TraceOrder = (typeof TraceOrder)[keyof typeof TraceOrder];

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

const TraceDetailTab = {
  Overview: "overview",
  Attributes: "attributes",
  Events: "events",
  Logs: "logs",
} as const;
export type TraceDetailTab =
  (typeof TraceDetailTab)[keyof typeof TraceDetailTab];

const SpanEventType = {
  Error: "error",
  Info: "info",
} as const;
export type SpanEventType = (typeof SpanEventType)[keyof typeof SpanEventType];

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
