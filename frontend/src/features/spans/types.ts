import type { SharedTime } from "@/shared/composables/useGlobalTimeRange";

export type SpanStatusFilter = "all" | "error" | "ok";

export interface SpansFilters extends SharedTime {
  search: string;
  status: SpanStatusFilter;
  service: string;
}

export interface SpansListParams extends SpansFilters {
  cursor?: string;
}

export const DEFAULT_SPANS_FILTERS: SpansFilters = {
  search: "",
  status: "all",
  service: "all",
};
