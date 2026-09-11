import type { SpanKind } from "@/shared/types/entities/span";

const SpanStatusFilter = {
  All: "all",
  Error: "error",
  Ok: "ok",
} as const;
export type SpanStatusFilter =
  (typeof SpanStatusFilter)[keyof typeof SpanStatusFilter];

export interface SpansFilters {
  search: string;
  status: SpanStatusFilter;
  kind: "all" | SpanKind;
  service: string;
  start?: string;
  end?: string;
}

export interface SpansListParams extends SpansFilters {
  cursor?: string;
}

export const DEFAULT_SPANS_FILTERS: SpansFilters = {
  search: "",
  status: "all",
  kind: "all",
  service: "all",
};
